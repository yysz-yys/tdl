# TDL Daemon & Cross-Platform Dashboard Architecture

This module `app/daemon` implements the core background service (Daemon) for the TDL cross-platform visual dashboard.

## Architecture

The system uses a **Local Daemon + Web UI** separation architecture:
1. **Go Daemon**: Runs a local HTTP Server and WebSocket Hub.
2. **Web UI**: A Vue 3 / React frontend that connects to the Daemon via REST API and WebSocket.
3. **Cross-Platform Shells**:
   - **Windows**: Wails or standard Web browser.
   - **Android**: Gomobile AAR running in an Android Foreground Service, with a WebView or Native UI.

## Components

- `server.go`: Sets up the `gorilla/mux` HTTP router and REST API endpoints.
- `ws.go`: Implements the WebSocket Hub using `coder/websocket` to push real-time download/upload progress to the UI.
- `task.go`: The asynchronous Task Manager that replaces the original blocking CLI download logic.
- `cmd/mobile/mobile.go`: The Gomobile export entry point for Android.
- `cmd/daemon.go`: The CLI command `tdl daemon` to start the standalone service.

## Next Steps for Frontend

### 1. Windows (Wails)
To package this as a Windows Desktop application:
```bash
wails init -n tdl-desktop -t vue-ts
# Replace the wails Go backend with our app/daemon/server.go
wails build -platform windows/amd64
```

### 2. Android (Gomobile)
To compile the core engine into an Android Library (`.aar`):
```bash
gomobile bind -target=android -androidapi 21 ./cmd/mobile
```
Then import `mobile.aar` into Android Studio, start the engine in a `Service`:
```java
import mobile.Mobile;

public class TdlService extends Service {
    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        new Thread(() -> {
            try {
                Mobile.startEngine(8080);
            } catch (Exception e) {
                e.printStackTrace();
            }
        }).start();
        return START_STICKY;
    }
}
```
