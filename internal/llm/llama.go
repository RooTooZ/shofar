package llm

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/llama.cpp/include -I${SRCDIR}/../../third_party/llama.cpp/ggml/include
// Note: LDFLAGS are set via Makefile's CGO_LDFLAGS to avoid ggml conflicts with whisper.cpp

#include <stdlib.h>
#include "llama.h"

// Helper function to create default model params
static struct llama_model_params get_default_model_params() {
    return llama_model_default_params();
}

// Helper function to create default context params
static struct llama_context_params get_default_context_params() {
    return llama_context_default_params();
}
*/
import "C"
import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// LlamaModel represents a loaded llama.cpp model.
type LlamaModel struct {
	mu      sync.Mutex
	model   *C.struct_llama_model
	ctx     *C.struct_llama_context
	sampler *C.struct_llama_sampler
	nCtx    int
}

// NewLlamaModel loads a GGUF model from file.
func NewLlamaModel(modelPath string, nCtx int) (*LlamaModel, error) {
	if nCtx <= 0 {
		nCtx = 2048
	}

	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	// Model params
	mparams := C.get_default_model_params()

	model := C.llama_model_load_from_file(cPath, mparams)
	if model == nil {
		return nil, errors.New("failed to load model")
	}

	// Context params
	cparams := C.get_default_context_params()
	cparams.n_ctx = C.uint32_t(nCtx)
	cparams.n_batch = C.uint32_t(512)

	ctx := C.llama_init_from_model(model, cparams)
	if ctx == nil {
		C.llama_model_free(model)
		return nil, errors.New("failed to create context")
	}

	// Create sampler chain
	sparams := C.llama_sampler_chain_default_params()
	sampler := C.llama_sampler_chain_init(sparams)

	// Add samplers: temp -> top_k -> top_p -> greedy
	C.llama_sampler_chain_add(sampler, C.llama_sampler_init_temp(0.1))
	C.llama_sampler_chain_add(sampler, C.llama_sampler_init_top_k(40))
	C.llama_sampler_chain_add(sampler, C.llama_sampler_init_top_p(0.9, 1))
	C.llama_sampler_chain_add(sampler, C.llama_sampler_init_dist(C.LLAMA_DEFAULT_SEED))

	return &LlamaModel{
		model:   model,
		ctx:     ctx,
		sampler: sampler,
		nCtx:    nCtx,
	}, nil
}

// Generate generates text completion for the given prompt.
func (m *LlamaModel) Generate(prompt string, maxTokens int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.model == nil || m.ctx == nil {
		return "", errors.New("model not loaded")
	}

	if maxTokens <= 0 {
		maxTokens = 256
	}

	// Tokenize prompt
	tokens, err := m.tokenize(prompt, true)
	if err != nil {
		return "", err
	}

	if len(tokens) == 0 {
		return "", errors.New("empty prompt")
	}

	// Clear memory (KV cache)
	mem := C.llama_get_memory(m.ctx)
	C.llama_memory_clear(mem, C.bool(true))

	// Create batch
	batch := C.llama_batch_get_one((*C.llama_token)(&tokens[0]), C.int32_t(len(tokens)))

	// Decode prompt
	if C.llama_decode(m.ctx, batch) != 0 {
		return "", errors.New("failed to decode prompt")
	}

	// Generate tokens
	var result strings.Builder
	nCur := len(tokens)

	for i := 0; i < maxTokens; i++ {
		// Sample next token
		newToken := C.llama_sampler_sample(m.sampler, m.ctx, -1)

		// Check for EOS
		if C.llama_vocab_is_eog(C.llama_model_get_vocab(m.model), newToken) {
			break
		}

		// Convert token to text
		piece := m.tokenToPiece(newToken)
		result.WriteString(piece)

		// Prepare batch for next token
		batch = C.llama_batch_get_one(&newToken, 1)

		// Decode
		if C.llama_decode(m.ctx, batch) != 0 {
			break
		}

		nCur++
		if nCur >= m.nCtx {
			break
		}
	}

	return result.String(), nil
}

// tokenize converts text to tokens.
func (m *LlamaModel) tokenize(text string, addBos bool) ([]C.llama_token, error) {
	vocab := C.llama_model_get_vocab(m.model)

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	// Estimate token count
	nTokens := len(text) + 16
	tokens := make([]C.llama_token, nTokens)

	bos := C.bool(addBos)
	special := C.bool(true)

	n := C.llama_tokenize(vocab, cText, C.int32_t(len(text)),
		(*C.llama_token)(&tokens[0]), C.int32_t(nTokens), bos, special)

	if n < 0 {
		// Need more space
		nTokens = int(-n)
		tokens = make([]C.llama_token, nTokens)
		n = C.llama_tokenize(vocab, cText, C.int32_t(len(text)),
			(*C.llama_token)(&tokens[0]), C.int32_t(nTokens), bos, special)
	}

	if n < 0 {
		return nil, errors.New("tokenization failed")
	}

	return tokens[:n], nil
}

// tokenToPiece converts a token to text.
func (m *LlamaModel) tokenToPiece(token C.llama_token) string {
	vocab := C.llama_model_get_vocab(m.model)

	buf := make([]byte, 64)
	n := C.llama_token_to_piece(vocab, token, (*C.char)(unsafe.Pointer(&buf[0])), C.int32_t(len(buf)), 0, C.bool(true))

	if n < 0 {
		// Need more space
		buf = make([]byte, -n)
		n = C.llama_token_to_piece(vocab, token, (*C.char)(unsafe.Pointer(&buf[0])), C.int32_t(len(buf)), 0, C.bool(true))
	}

	if n <= 0 {
		return ""
	}

	return string(buf[:n])
}

// Close frees the model resources.
func (m *LlamaModel) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sampler != nil {
		C.llama_sampler_free(m.sampler)
		m.sampler = nil
	}

	if m.ctx != nil {
		C.llama_free(m.ctx)
		m.ctx = nil
	}

	if m.model != nil {
		C.llama_model_free(m.model)
		m.model = nil
	}
}

// CorrectText исправляет текст с помощью LLM.
func (m *LlamaModel) CorrectText(ctx context.Context, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}

	// Формируем промпт для коррекции
	prompt := fmt.Sprintf(`<|im_start|>system
Ты помощник для исправления ошибок распознавания речи. Исправь ошибки и расставь знаки препинания. Верни только исправленный текст без пояснений.<|im_end|>
<|im_start|>user
%s<|im_end|>
<|im_start|>assistant
`, text)

	// Проверяем контекст
	select {
	case <-ctx.Done():
		return text, ctx.Err()
	default:
	}

	result, err := m.Generate(prompt, 256)
	if err != nil {
		return text, fmt.Errorf("llm generate: %w", err)
	}

	// Очищаем результат от лишнего
	corrected := strings.TrimSpace(result)

	// Убираем возможные теги в конце
	if idx := strings.Index(corrected, "<|im_end|>"); idx != -1 {
		corrected = strings.TrimSpace(corrected[:idx])
	}

	return corrected, nil
}
