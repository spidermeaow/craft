# Rev.40 candidate: Windows window icon

`gui.setWindowIcon(window, path)` accepts a Windows `.ico` file. Call it after `gui.openWindow`; a later call replaces the same window's icon. Relative paths resolve from the process working directory. A missing or invalid file raises `UiArgumentError` and leaves the previous icon in place.

```craft
import gui "gui"

func main() {
    let window: UiWindow = gui.openWindow("Example", 900, 600)
    gui.setWindowIcon(window, "assets/app.ico")
    gui.mount(window, gui.panel("root", [gui.text("hello", "Hello")]))
    var running: Bool = true
    while running {
        if let event = gui.nextEvent(-1) {
            if event.kind == "CloseRequested" { running = false }
        }
    }
    gui.closeWindow(window)
}
```

Runtime sets both `ICON_BIG` and `ICON_SMALL` with `WM_SETICON`, using the window DPI at load time. It destroys handles that it owns when the icon changes or the window closes. The default icon remains if this API is never called. The API changes the running window; it does not embed an icon in an executable or installer.

## Manual check in UI Showcase

From `D:\02-Repository\app`, run `craft package lock` after obtaining the matching 0.1.40 toolchain, then `craft check` and `craft run`. Open **Window Icon**. Compare the Blue and Orange buttons in the title bar, Alt+Tab, and Taskbar. Click **Missing file** and confirm the status reports an error while the previous icon remains. Switch several times and close the window; inspect handle usage if desired. Repeat at another Windows DPI setting. Results are pending user testing.
