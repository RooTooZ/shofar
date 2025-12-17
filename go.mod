module shofar

go 1.23.8

require (
	gioui.org v0.9.0
	github.com/gen2brain/beeep v0.0.0-20240516210008-9c006672e7f4
	github.com/getlantern/systray v1.2.2
	github.com/ggerganov/whisper.cpp/bindings/go v0.0.0-00010101000000-000000000000
	github.com/gordonklaus/portaudio v0.0.0-20230709114228-aafa478834f5
	github.com/ncruces/zenity v0.10.14
	golang.design/x/hotkey v0.4.1
)

require (
	gioui.org/shader v1.0.8 // indirect
	github.com/akavel/rsrc v0.10.2 // indirect
	github.com/alphacep/vosk-api/go v0.3.50 // indirect
	github.com/dchest/jsmin v0.0.0-20220218165748-59f39799265f // indirect
	github.com/getlantern/context v0.0.0-20190109183933-c447772a6520 // indirect
	github.com/getlantern/errors v0.0.0-20190325191628-abdb3e3e36f7 // indirect
	github.com/getlantern/golog v0.0.0-20190830074920-4ef2e798c2d7 // indirect
	github.com/getlantern/hex v0.0.0-20190417191902-c6586a6fe0b7 // indirect
	github.com/getlantern/hidden v0.0.0-20190325191715-f02dbb02be55 // indirect
	github.com/getlantern/ops v0.0.0-20190325191751-d70cb0d6f85f // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/go-text/typesetting v0.3.0 // indirect
	github.com/go-toast/toast v0.0.0-20190211030409-01e6764cf0a4 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/josephspurrier/goversioninfo v1.4.1 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/randall77/makefat v0.0.0-20210315173500-7ddd0e42c844 // indirect
	github.com/tadvi/systray v0.0.0-20190226123456-11a2b8fa57af // indirect
	golang.design/x/mainthread v0.3.0 // indirect
	golang.org/x/exp/shiny v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/image v0.26.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)

replace github.com/ggerganov/whisper.cpp/bindings/go => ./third_party/whisper.cpp/bindings/go
