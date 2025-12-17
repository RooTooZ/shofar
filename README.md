<div align="center">

<img src="assets/logo.png" alt="Shofar Logo" width="128">

# Shofar

### Cross-Platform Voice-to-Text

**Speak naturally. Type instantly.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Linux](https://img.shields.io/badge/Linux-FCC624?style=flat-square&logo=linux&logoColor=black)](https://www.linux.org/)
[![macOS](https://img.shields.io/badge/macOS-000000?style=flat-square&logo=apple&logoColor=white)](https://www.apple.com/macos/)
[![Windows](https://img.shields.io/badge/Windows-0078D6?style=flat-square&logo=windows&logoColor=white)](https://www.microsoft.com/windows/)

[Features](#-features) • [Installation](#-installation) • [Usage](#-usage) • [Models](#-models) • [Architecture](#-architecture)

---

```
    ┌─────────────────────────────────┐
    │       🎙 Recording...           │
    │    ▁▂▃▅▆▇█▇▆▅▃▂▁▂▃▅▆▇█▇▆▅     │
    │            0:03                 │
    └─────────────────────────────────┘
              ⬇️ Press hotkey
    ┌─────────────────────────────────┐
    │  📝 Result                      │
    │  ┌───────────────────────────┐  │
    │  │ Ваш распознанный текст    │  │
    │  │ появится здесь            │  │
    │  └───────────────────────────┘  │
    │     [Вставить]    [Копировать]  │
    └─────────────────────────────────┘
```

*Press hotkey → Speak → Text appears in any application*

</div>

---

## 💡 Why Shofar?

Most voice input solutions are either **cloud-based** (privacy concerns), **require complex setup**, or **don't integrate** with the desktop. Shofar is different:

| Problem | Shofar Solution |
|---------|-----------------|
| 🔐 Privacy concerns | **100% Offline** — voice never leaves your machine |
| 📦 Complex dependencies | **Single Binary** — no Python, Docker, or npm |
| 🖥️ Limited integration | **System-wide** — works in any app, any text field |
| ⚡ Speed vs accuracy trade-off | **Hot-swap Models** — switch engines on the fly |

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🎯 Core
- **Push-to-talk** with customizable hotkey
- **Real-time waveform** visualization
- **Auto-insert** into active window
- **System tray** with status indication

</td>
<td width="50%">

### 🧠 Engines
- **Whisper** (OpenAI) — best accuracy
- **Vosk** — fastest, lowest latency
- **LLM correction** — fix recognition errors

</td>
</tr>
<tr>
<td>

### 🌍 Languages
- 🇷🇺 Russian (primary)
- 🇬🇧 English
- 🔄 Auto-detection

</td>
<td>

### 🎨 Interface
- Dark theme floating window
- Editable results before insert
- Copy to clipboard option
- Desktop notifications

</td>
</tr>
</table>

---

## 📦 Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/RooTooZ/shofar/master/install.sh | sh
```

Or download from [Releases](https://github.com/RooTooZ/shofar/releases).

---

### Build from Source

#### 1. Install Dependencies

<details>
<summary><b>🐧 Arch Linux</b></summary>

```bash
sudo pacman -S portaudio libayatana-appindicator gtk3 pkg-config cmake xdotool
```

For Wayland:
```bash
sudo pacman -S wtype
```
</details>

<details>
<summary><b>🐧 Ubuntu / Debian</b></summary>

```bash
sudo apt install gcc pkg-config cmake xdotool \
    libportaudio2 portaudio19-dev \
    libayatana-appindicator3-dev libgtk-3-dev \
    libwayland-dev libx11-dev libx11-xcb-dev \
    libxkbcommon-x11-dev libgles2-mesa-dev \
    libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev
```

For Wayland:
```bash
sudo apt install wtype
```
</details>

<details>
<summary><b>🍎 macOS</b></summary>

```bash
brew install portaudio pkg-config cmake
```
</details>

<details>
<summary><b>🪟 Windows</b></summary>

Requires:
- MinGW-w64 or MSYS2
- CMake
- PortAudio

```powershell
# Using MSYS2
pacman -S mingw-w64-x86_64-portaudio mingw-w64-x86_64-cmake
```
</details>

#### 2. Build

```bash
git clone https://github.com/RooTooZ/shofar.git
cd shofar

# Build everything (whisper.cpp + llama.cpp + binary)
make all

# Install to ~/.local/bin
make install
```

#### 3. Download a Model

Models download automatically via Settings UI, or:

```bash
make download-model-tiny    # 75 MB  — fast
make download-model-small   # 466 MB — balanced
make download-model-turbo   # 574 MB — best quality
```

---

## 🚀 Usage

```bash
shofar    # or ./bin/shofar
```

### Quick Start

| Step | Action |
|:----:|--------|
| **1** | Press **Ctrl+Shift+Space** |
| **2** | 🎙️ Speak while window is visible |
| **3** | Press hotkey again to stop |
| **4** | ✏️ Edit text if needed |
| **5** | **Enter** = insert, **Esc** = cancel |

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl+Shift+Space` | Start / Stop recording |
| `Enter` | Insert text into active window |
| `Esc` | Cancel and close |

### Tray Menu

Right-click tray icon for:
- ⚙️ **Settings** — models, hotkey, language
- 🔔 **Notifications** — toggle on/off
- ❌ **Quit**

---

## 🧠 Models

### Speech Recognition

| Engine | Model | Size | Speed | Accuracy | Best For |
|--------|-------|------|:-----:|:--------:|----------|
| Whisper | `tiny` | 75 MB | ⚡⚡⚡ | ★★☆☆ | Quick commands |
| Whisper | `small` | 466 MB | ⚡⚡ | ★★★☆ | Daily use |
| Whisper | `turbo` | 574 MB | ⚡ | ★★★★ | Documents |
| Vosk | `small-ru` | 45 MB | ⚡⚡⚡⚡ | ★★☆☆ | Real-time |

### LLM Text Correction (Optional)

| Model | Size | Description |
|-------|------|-------------|
| Qwen2.5-1.5B | 1.1 GB | Fixes punctuation & typos |

---

## 🏗 Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                           SHOFAR                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐ │
│   │  Hotkey  │ -> │  Audio   │ -> │  Speech  │ -> │   Input  │ │
│   │  Handler │    │ Recorder │    │  Engine  │    │ Emulator │ │
│   └──────────┘    └──────────┘    └──────────┘    └──────────┘ │
│        │                              │                         │
│        v                              v                         │
│   ┌──────────┐                  ┌──────────┐                    │
│   │   Tray   │                  │   LLM    │                    │
│   │   Icon   │                  │ Corrector│                    │
│   └──────────┘                  └──────────┘                    │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  whisper.cpp  │  llama.cpp  │  vosk-api  │  PortAudio  │  Gio  │
└─────────────────────────────────────────────────────────────────┘
```

### Project Structure

```
shofar/
├── cmd/shofar/            # Entry point
├── internal/
│   ├── app/               # Main coordinator
│   ├── audio/             # PortAudio recording
│   ├── speech/            # Engine factory (Whisper/Vosk)
│   ├── llm/               # LLM text correction
│   ├── hotkey/            # Global hotkey
│   ├── tray/              # System tray
│   ├── waveform/          # Recording UI
│   ├── settings/          # Settings UI
│   ├── input/             # xdotool/wtype wrapper
│   └── i18n/              # Translations
└── third_party/           # whisper.cpp, llama.cpp, vosk
```

### Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ (CGO) |
| Speech | [whisper.cpp](https://github.com/ggerganov/whisper.cpp), [Vosk](https://alphacephei.com/vosk/) |
| LLM | [llama.cpp](https://github.com/ggerganov/llama.cpp) |
| Audio | [PortAudio](http://www.portaudio.com/) |
| GUI | [Gio](https://gioui.org/) |
| Tray | [systray](https://github.com/getlantern/systray) |
| Text Input | xdotool (X11) / wtype (Wayland) / CGEventPost (macOS) / SendInput (Windows) |

---

## ⚙️ Configuration

Config stored at `~/.local/bin/config.json`:

```json
{
  "model_id": "whisper-small",
  "language": "ru",
  "hotkey": { "key": "Space", "modifiers": ["Ctrl", "Shift"] },
  "llm_enabled": false,
  "notifications": true
}
```

---

## 🛠 Development

```bash
make whisper-lib    # Build whisper.cpp
make llama-lib      # Build llama.cpp
make build          # Build application
make run            # Run for testing
make clean          # Clean artifacts
```

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push branch (`git push origin feature/amazing`)
5. Open Pull Request

---

## 📄 License

[MIT License](LICENSE) — use freely, attribution appreciated.

---

<div align="center">

**[⬆ Back to top](#-shofar)**

Made with ❤️ for desktop users everywhere

*Star ⭐ if you find it useful!*

</div>
