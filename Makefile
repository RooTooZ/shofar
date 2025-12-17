.PHONY: build clean clean-all generate-icons whisper-lib llama-lib download-vosk-lib \
	all install run deps check help release release-linux release-darwin release-windows

VERSION := 0.1.0
LDFLAGS := -s -w -X main.Version=$(VERSION)
BIN_DIR := bin
WHISPER_DIR := third_party/whisper.cpp
LLAMA_DIR := third_party/llama.cpp

# === Vosk ===
VOSK_VERSION := 0.3.45
VOSK_DIR := third_party/vosk
VOSK_LIB_DIR := $(CURDIR)/$(VOSK_DIR)/vosk-linux-x86_64-$(VOSK_VERSION)

# Пути к whisper.cpp библиотеке (только libwhisper, ggml берём из llama.cpp)
WHISPER_CFLAGS := -I$(CURDIR)/$(WHISPER_DIR)/include -I$(CURDIR)/$(WHISPER_DIR)/ggml/include
WHISPER_LDFLAGS := -L$(CURDIR)/$(WHISPER_DIR)/build/src -lwhisper

# Пути к llama.cpp библиотеке (ggml используется общий)
LLAMA_CFLAGS := -I$(CURDIR)/$(LLAMA_DIR)/include -I$(CURDIR)/$(LLAMA_DIR)/ggml/include
LLAMA_LDFLAGS := -L$(CURDIR)/$(LLAMA_DIR)/build/src -lllama

# Общие ggml библиотеки (используем из llama.cpp - новее)
GGML_LDFLAGS := -L$(CURDIR)/$(LLAMA_DIR)/build/ggml/src -lggml -lggml-base -lggml-cpu -lstdc++ -lm -fopenmp

# Пути к Vosk библиотеке
VOSK_CFLAGS := -I$(VOSK_LIB_DIR)
VOSK_LDFLAGS := -L$(VOSK_LIB_DIR) -lvosk -Wl,-rpath,$(VOSK_LIB_DIR)

# Объединённые CGO флаги (Whisper + Vosk + Llama + общий ggml)
export CGO_CFLAGS := $(WHISPER_CFLAGS) $(VOSK_CFLAGS) $(LLAMA_CFLAGS)
export CGO_LDFLAGS := $(WHISPER_LDFLAGS) $(LLAMA_LDFLAGS) $(GGML_LDFLAGS) $(VOSK_LDFLAGS)
export CGO_CFLAGS_ALLOW := -mfma|-mf16c|-mavx|-mavx2
export C_INCLUDE_PATH := $(CURDIR)/$(WHISPER_DIR)/include:$(CURDIR)/$(WHISPER_DIR)/ggml/include:$(CURDIR)/$(LLAMA_DIR)/include
export LIBRARY_PATH := $(CURDIR)/$(WHISPER_DIR)/build/src:$(CURDIR)/$(WHISPER_DIR)/build/ggml/src:$(CURDIR)/$(LLAMA_DIR)/build/src

# Генерация иконок
generate-icons:
	@mkdir -p embedded
	@echo "Генерация иконок..."
	go run scripts/generate_icons.go

# Клонирование и сборка whisper.cpp
whisper-lib:
	@if [ ! -d $(WHISPER_DIR) ]; then \
		echo "Клонирование whisper.cpp..."; \
		mkdir -p third_party; \
		git clone --depth 1 https://github.com/ggml-org/whisper.cpp.git $(WHISPER_DIR); \
	fi
	@if [ ! -f $(WHISPER_DIR)/build/src/libwhisper.a ]; then \
		echo "Сборка whisper.cpp..."; \
		cd $(WHISPER_DIR) && \
		cmake -B build -DBUILD_SHARED_LIBS=OFF -DWHISPER_BUILD_EXAMPLES=OFF -DWHISPER_BUILD_TESTS=OFF && \
		cmake --build build --config Release -j$$(nproc); \
	fi
	@echo "whisper.cpp готов"

# Клонирование и сборка llama.cpp
llama-lib:
	@if [ ! -d $(LLAMA_DIR) ]; then \
		echo "Клонирование llama.cpp..."; \
		mkdir -p third_party; \
		git clone --depth 1 https://github.com/ggml-org/llama.cpp.git $(LLAMA_DIR); \
	fi
	@if [ ! -f $(LLAMA_DIR)/build/src/libllama.a ]; then \
		echo "Сборка llama.cpp..."; \
		cd $(LLAMA_DIR) && \
		cmake -B build -DBUILD_SHARED_LIBS=OFF -DLLAMA_BUILD_TOOLS=OFF -DLLAMA_BUILD_TESTS=OFF -DLLAMA_BUILD_COMMON=OFF && \
		cmake --build build --target llama --config Release -j$$(nproc); \
	fi
	@echo "llama.cpp готов"

# Скачивание библиотеки Vosk
download-vosk-lib:
	@if [ -d $(VOSK_LIB_DIR) ]; then \
		echo "Vosk библиотека уже есть"; \
	else \
		mkdir -p $(VOSK_DIR); \
		echo "Скачивание Vosk библиотеки v$(VOSK_VERSION)..."; \
		wget -q --show-progress -O /tmp/vosk-linux.zip \
			"https://github.com/alphacep/vosk-api/releases/download/v$(VOSK_VERSION)/vosk-linux-x86_64-$(VOSK_VERSION).zip"; \
		unzip -o /tmp/vosk-linux.zip -d $(VOSK_DIR); \
		rm /tmp/vosk-linux.zip; \
		echo "Vosk библиотека установлена в $(VOSK_LIB_DIR)"; \
	fi

# Проверка иконок
check-icons:
	@test -f embedded/icon_idle.png || $(MAKE) generate-icons

# Единая сборка (Whisper + Vosk + Llama)
build: whisper-lib llama-lib download-vosk-lib check-icons
	@mkdir -p $(BIN_DIR)/models/whisper
	@mkdir -p $(BIN_DIR)/models/vosk
	@mkdir -p $(BIN_DIR)/models/llm
	CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/shofar ./cmd/shofar
	@echo ""
	@echo "Собрано: $(BIN_DIR)/shofar"
	@echo "Размер: $$(du -h $(BIN_DIR)/shofar | cut -f1)"
	@echo ""
	@echo "Модели будут скачиваться в $(BIN_DIR)/models/ при выборе в настройках"

# Полная сборка с нуля
all: whisper-lib llama-lib download-vosk-lib generate-icons build

# Установка
install: build
	@mkdir -p ~/.local/bin
	@mkdir -p ~/.local/share/shofar/models/whisper
	@mkdir -p ~/.local/share/shofar/models/vosk
	@mkdir -p ~/.local/share/shofar/models/llm
	cp $(BIN_DIR)/shofar ~/.local/bin/shofar
	@echo "Установлено в ~/.local/bin/shofar"
	@echo "Модели будут храниться в ~/.local/share/shofar/models/"

# Очистка
clean:
	rm -rf $(BIN_DIR)/

clean-all: clean
	rm -rf $(WHISPER_DIR)
	rm -rf $(LLAMA_DIR)
	rm -rf $(VOSK_DIR)

# Загрузка зависимостей
deps:
	go mod download
	go mod tidy

# Проверка кода
check:
	go vet ./...

# Запуск
run: build
	LD_LIBRARY_PATH=$(VOSK_LIB_DIR):$$LD_LIBRARY_PATH ./$(BIN_DIR)/shofar

# === Release ===
RELEASE_DIR := release

release-linux: whisper-lib llama-lib download-vosk-lib check-icons
	@mkdir -p $(RELEASE_DIR)
	@echo "Building for Linux amd64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
		go build -ldflags "$(LDFLAGS)" -o $(RELEASE_DIR)/shofar-linux-amd64 ./cmd/shofar
	@cd $(RELEASE_DIR) && tar -czvf shofar-$(VERSION)-linux-amd64.tar.gz shofar-linux-amd64
	@rm $(RELEASE_DIR)/shofar-linux-amd64
	@echo "Created: $(RELEASE_DIR)/shofar-$(VERSION)-linux-amd64.tar.gz"

release: release-linux
	@echo ""
	@echo "Release files in $(RELEASE_DIR)/"
	@ls -lh $(RELEASE_DIR)/

# Помощь
help:
	@echo "Shofar - голосовой ввод текста"
	@echo ""
	@echo "=== Единый бинарник с поддержкой Whisper и Vosk ==="
	@echo ""
	@echo "Теперь приложение поддерживает оба движка в одном бинарнике!"
	@echo "Модели скачиваются через UI настроек при первом использовании."
	@echo ""
	@echo "=== Доступные модели ==="
	@echo ""
	@echo "Whisper (квантизированные, рекомендуется):"
	@echo "  - Tiny Q5     32 MB   - самая быстрая"
	@echo "  - Base Q5     60 MB   - хороший баланс"
	@echo "  - Small Q5   190 MB   - лучше качество"
	@echo "  - Turbo Q5   574 MB   - быстрая + отличное качество"
	@echo ""
	@echo "Whisper (оригинальные FP16):"
	@echo "  - Tiny        75 MB"
	@echo "  - Base       142 MB"
	@echo "  - Small      466 MB"
	@echo ""
	@echo "Vosk:"
	@echo "  - Russian   1.8 GB   - офлайн, только русский"
	@echo ""
	@echo "=== Команды ==="
	@echo ""
	@echo "  make build          - Собрать приложение"
	@echo "  make run            - Собрать и запустить"
	@echo "  make install        - Установить в ~/.local/bin"
	@echo ""
	@echo "=== Подготовка ==="
	@echo ""
	@echo "  make whisper-lib      - Клонировать и собрать whisper.cpp"
	@echo "  make download-vosk-lib - Скачать библиотеку Vosk"
	@echo "  make generate-icons   - Сгенерировать иконки"
	@echo "  make all              - Полная сборка с нуля"
	@echo ""
	@echo "=== Очистка ==="
	@echo ""
	@echo "  make clean          - Удалить бинарники"
	@echo "  make clean-all      - Удалить всё (включая библиотеки)"
	@echo ""
	@echo "=== Использование ==="
	@echo ""
	@echo "1. Запустите приложение: make run"
	@echo "2. Кликните на иконку в трее -> 'Модель распознавания...'"
	@echo "3. Выберите движок (Whisper/Vosk) и модель"
	@echo "4. Нажмите 'Скачать' если модель не загружена"
	@echo "5. Нажмите 'Применить' для активации модели"
