//go:build windows

package main

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const (
	cwUseDefault        = 0x80000000
	wsOverlapped        = 0x00000000
	wsCaption           = 0x00C00000
	wsSysMenu           = 0x00080000
	wsThickFrame        = 0x00040000
	wsMinimizeBox       = 0x00020000
	wsMaximizeBox       = 0x00010000
	wsVisible           = 0x10000000
	wsClipChildren      = 0x02000000
	launcherWindowStyle = wsOverlapped | wsCaption | wsSysMenu | wsThickFrame | wsMinimizeBox | wsMaximizeBox | wsVisible | wsClipChildren
	wsChild             = 0x40000000
	wsExClientEdge      = 0x00000200
	bsPushButton        = 0x00000000
	bsOwnerDraw         = 0x0000000B
	esMultiline         = 0x0004
	esAutovScroll       = 0x0040
	esReadOnly          = 0x0800
	wsVScroll           = 0x00200000
	wmDestroy           = 0x0002
	wmCreate            = 0x0001
	wmSize              = 0x0005
	wmPaint             = 0x000F
	wmGetMinMaxInfo     = 0x0024
	wmCtlColorEdit      = 0x0133
	wmCtlColorStatic    = 0x0138
	wmCommand           = 0x0111
	wmDrawItem          = 0x002B
	wmAppLog            = 0x8001
	wmDpichanged        = 0x02E0
	wmSetIcon           = 0x0080
	wmSetText           = 0x000C
	wmGetTextLen        = 0x000E
	wmSetFont           = 0x0030
	emSetSel            = 0x00B1
	emReplaceSel        = 0x00C2
	emScrollCaret       = 0x00B7
	odsSelected         = 0x0001
	iconSmall           = 0
	iconBig             = 1
	colorWindow         = 5
	swShow              = 5
	idOpenDocs          = 1001
	idOpenFolder        = 1002
	idCopyURL           = 1003
)

type hwnd uintptr
type hinstance uintptr
type hicon uintptr
type hfont uintptr
type hbrush uintptr
type hdc uintptr
type hpen uintptr

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type paintStruct struct {
	hdc       hdc
	erase     int32
	rcPaint   rect
	restore   int32
	incUpdate int32
	reserved  [32]byte
}

type point struct {
	x int32
	y int32
}

type minMaxInfo struct {
	reserved     point
	maxSize      point
	maxPosition  point
	minTrackSize point
	maxTrackSize point
}

type wndClassEx struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   hinstance
	icon       hicon
	cursor     uintptr
	background hbrush
	menuName   *uint16
	className  *uint16
	iconSm     hicon
}

type msg struct {
	hwnd    hwnd
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	ptX     int32
	ptY     int32
}

type drawItemStruct struct {
	ctlType    uint32
	ctlID      uint32
	itemID     uint32
	itemAction uint32
	itemState  uint32
	hwndItem   hwnd
	hdc        hdc
	rcItem     rect
	itemData   uintptr
}

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	gdi32             = syscall.NewLazyDLL("gdi32.dll")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procRegisterClass = user32.NewProc("RegisterClassExW")
	procCreateWindow  = user32.NewProc("CreateWindowExW")
	procDefWindowProc = user32.NewProc("DefWindowProcW")
	procAdjustRect    = user32.NewProc("AdjustWindowRectEx")
	procShowWindow    = user32.NewProc("ShowWindow")
	procUpdateWindow  = user32.NewProc("UpdateWindow")
	procGetMessage    = user32.NewProc("GetMessageW")
	procTranslateMsg  = user32.NewProc("TranslateMessage")
	procDispatchMsg   = user32.NewProc("DispatchMessageW")
	procPostQuit      = user32.NewProc("PostQuitMessage")
	procPostMessage   = user32.NewProc("PostMessageW")
	procSendMessage   = user32.NewProc("SendMessageW")
	procSetWindowText = user32.NewProc("SetWindowTextW")
	procMoveWindow    = user32.NewProc("MoveWindow")
	procGetClientRect = user32.NewProc("GetClientRect")
	procInvalidate    = user32.NewProc("InvalidateRect")
	procBeginPaint    = user32.NewProc("BeginPaint")
	procEndPaint      = user32.NewProc("EndPaint")
	procFillRect      = user32.NewProc("FillRect")
	procDrawText      = user32.NewProc("DrawTextW")
	procSetTextColor  = gdi32.NewProc("SetTextColor")
	procSetBkMode     = gdi32.NewProc("SetBkMode")
	procSetBkColor    = gdi32.NewProc("SetBkColor")
	procCreateBrush   = gdi32.NewProc("CreateSolidBrush")
	procCreatePen     = gdi32.NewProc("CreatePen")
	procSelectObject  = gdi32.NewProc("SelectObject")
	procRoundRect     = gdi32.NewProc("RoundRect")
	procDeleteObject  = gdi32.NewProc("DeleteObject")
	procLoadIcon      = user32.NewProc("LoadIconW")
	procLoadCursor    = user32.NewProc("LoadCursorW")
	procGetDpiForWin  = user32.NewProc("GetDpiForWindow")
	procSetDpiCtx     = user32.NewProc("SetProcessDpiAwarenessContext")
	procSetDpiAware   = user32.NewProc("SetProcessDPIAware")
	procCreateFont    = gdi32.NewProc("CreateFontW")
	procGetModule     = kernel32.NewProc("GetModuleHandleW")
	procExtractIcon   = shell32.NewProc("ExtractIconW")

	appURL       string
	appBaseDir   string
	appStop      func()
	mainWindow   hwnd
	logEdit      hwnd
	statusStatic hwnd
	statusPort   hwnd
	urlStatic    hwnd
	titleMain    hwnd
	descMain     hwnd
	addressLabel hwnd
	openButton   hwnd
	folderButton hwnd
	copyButton   hwnd
	logTitle     hwnd
	footerStatic hwnd
	defaultFont  hfont
	titleFont    hfont
	brandFont    hfont
	monoFont     hfont
	bgBrush      hbrush
	sideBrush    hbrush
	cardBrush    hbrush
	logBrush     hbrush
	whiteBrush   hbrush
	uiDPI        int32 = 96
	logMu        sync.Mutex
	pendingLogs  []string
	stopOnce     sync.Once
)

func runLauncher(baseDir, url string, logs <-chan string, stop func()) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	enableDPIAwareness()

	appURL = url
	appBaseDir = baseDir
	appStop = stop

	instance, _, _ := procGetModule.Call(0)
	className := syscall.StringToUTF16Ptr("MDOnlineLauncherWindow")
	title := syscall.StringToUTF16Ptr("MDOnline 本地文档服务")

	icon := loadAppIcon(baseDir)
	cursor, _, _ := procLoadCursor.Call(0, uintptr(32512))
	wc := wndClassEx{
		size:       uint32(unsafe.Sizeof(wndClassEx{})),
		wndProc:    syscall.NewCallback(wndProc),
		instance:   hinstance(instance),
		icon:       icon,
		cursor:     cursor,
		background: hbrush(0),
		className:  className,
		iconSm:     icon,
	}
	procRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))

	mainHwnd, _, _ := procCreateWindow.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		launcherWindowStyle,
		uintptr(cwUseDefault), uintptr(cwUseDefault), uintptr(adjustedWindowWidth(1120)), uintptr(adjustedWindowHeight(760)),
		0, 0, instance, 0,
	)
	mainWindow = hwnd(mainHwnd)
	procSendMessage.Call(mainHwnd, wmSetIcon, iconSmall, uintptr(icon))
	procSendMessage.Call(mainHwnd, wmSetIcon, iconBig, uintptr(icon))

	go func() {
		for line := range logs {
			appendLog(line)
		}
	}()

	procShowWindow.Call(mainHwnd, 1)
	procUpdateWindow.Call(mainHwnd)

	var m msg
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMsg.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMsg.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func wndProc(h hwnd, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		updateDPI(h)
		initTheme()
		createControls(h)
		layoutControls(h)
		return 0
	case wmSize:
		layoutControls(h)
		procInvalidate.Call(uintptr(h), 0, 1)
		return 0
	case wmDpichanged:
		updateDPI(h)
		return 0
	case wmPaint:
		paintWindow(h)
		return 0
	case wmGetMinMaxInfo:
		info := (*minMaxInfo)(unsafe.Pointer(lParam))
		info.minTrackSize.x = adjustedWindowWidth(1040)
		info.minTrackSize.y = adjustedWindowHeight(720)
		return 0
	case wmCtlColorStatic:
		if hwnd(lParam) == logEdit {
			procSetBkMode.Call(wParam, 2)
			procSetBkColor.Call(wParam, rgb(15, 23, 42))
			procSetTextColor.Call(wParam, rgb(203, 213, 225))
			return uintptr(logBrush)
		}
		procSetBkMode.Call(wParam, 1)
		switch hwnd(lParam) {
		case titleMain:
			procSetTextColor.Call(wParam, rgb(15, 23, 42))
			return uintptr(whiteBrush)
		case descMain, addressLabel, footerStatic, statusPort:
			procSetTextColor.Call(wParam, rgb(100, 116, 139))
		default:
			procSetTextColor.Call(wParam, rgb(23, 32, 51))
		}
		if isMainArea(hwnd(lParam)) {
			return uintptr(whiteBrush)
		}
		return uintptr(sideBrush)
	case wmCtlColorEdit:
		procSetBkMode.Call(wParam, 2)
		procSetBkColor.Call(wParam, rgb(15, 23, 42))
		procSetTextColor.Call(wParam, rgb(203, 213, 225))
		return uintptr(logBrush)
	case wmDestroy:
		requestStop()
		procPostQuit.Call(0)
		return 0
	case wmCommand:
		switch int(wParam & 0xffff) {
		case idOpenDocs:
			openBrowser(appURL)
		case idOpenFolder:
			openFolder(filepath.Join(appBaseDir, "docs"))
		case idCopyURL:
			copyURLToClipboard(appURL)
			setStatus("访问地址已复制")
		}
		return 0
	case wmDrawItem:
		drawButton((*drawItemStruct)(unsafe.Pointer(lParam)))
		return 1
	case wmAppLog:
		flushLogs()
		return 0
	}
	return defWindowProc(h, message, wParam, lParam)
}

func createControls(parent hwnd) {
	createStatic(parent, "M", 58, 52, 46, 48, createFont("Segoe UI", 31, 800))
	createStatic(parent, "MDOnline", 40, 120, 170, 36, brandFont)
	createStatic(parent, "本地 Markdown 文档服务", 40, 158, 210, 26, defaultFont)
	statusStatic = createStatic(parent, "●  服务运行中", 48, 222, 160, 24, defaultFont)
	statusPort = createStatic(parent, "端口 17621", 64, 250, 120, 22, defaultFont)

	titleMain = createStatic(parent, "本地文档服务已启动", 324, 66, 360, 38, titleFont)
	descMain = createStatic(parent, "文档服务已启动。需要访问时点击“打开文档”。", 324, 110, 520, 48, defaultFont)
	addressLabel = createStatic(parent, "访问地址", 324, 176, 90, 24, defaultFont)
	urlStatic = createStatic(parent, appURL, 344, 224, 360, 28, monoFont)

	openButton = createButton(parent, "打开文档", idOpenDocs, 800, 216, 118, 42)
	folderButton = createButton(parent, "打开 docs 文件夹", idOpenFolder, 324, 286, 156, 40)
	copyButton = createButton(parent, "复制访问地址", idCopyURL, 494, 286, 146, 40)

	logTitle = createStatic(parent, "运行日志", 324, 374, 90, 24, defaultFont)
	logEdit = createEdit(parent, "", 332, 412, 530, 150, monoFont)
	footerStatic = createStatic(parent, "关闭窗口将停止本地服务；最小化窗口时会保留在任务栏。", 324, 620, 520, 24, defaultFont)
}

func layoutControls(parent hwnd) {
	w, h := clientSize(parent)
	if w <= 0 || h <= 0 {
		return
	}
	left := int32(324)
	right := unscale(w) - 64
	if right < 920 {
		right = 920
	}
	cardW := right - left
	move(titleMain, left, 66, cardW-40, 38)
	move(descMain, left, 110, cardW-56, 48)
	move(addressLabel, left, 176, 90, 24)
	move(urlStatic, left+20, 224, cardW-178, 28)
	move(openButton, right-140, 216, 118, 42)
	move(folderButton, left, 286, 156, 40)
	move(copyButton, left+170, 286, 146, 40)

	logTop := int32(412)
	logBottom := unscale(h) - 116
	if logBottom < 590 {
		logBottom = 590
	}
	move(logTitle, left, 374, 90, 24)
	move(logEdit, left+12, logTop+12, cardW-24, logBottom-logTop-24)
	move(footerStatic, left, logBottom+20, cardW-20, 24)
}

func initTheme() {
	defaultFont = createFont("Segoe UI", 13, 400)
	titleFont = createFont("Segoe UI", 20, 700)
	brandFont = createFont("Segoe UI", 20, 800)
	monoFont = createFont("Cascadia Mono", 12, 400)
	bgBrush = createSolidBrush(244, 248, 251)
	sideBrush = createSolidBrush(234, 248, 251)
	cardBrush = createSolidBrush(255, 255, 255)
	logBrush = createSolidBrush(15, 23, 42)
	whiteBrush = createSolidBrush(255, 255, 255)
}

func paintWindow(parent hwnd) {
	var ps paintStruct
	dc, _, _ := procBeginPaint.Call(uintptr(parent), uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(uintptr(parent), uintptr(unsafe.Pointer(&ps)))
	clientW, clientH := clientSize(parent)
	w := unscale(clientW)
	h := unscale(clientH)
	if w < 920 {
		w = 920
	}
	if h < 600 {
		h = 600
	}
	right := w - 48
	if right < 936 {
		right = 936
	}
	logBottom := h - 116
	if logBottom < 590 {
		logBottom = 590
	}

	fill(dc, rect{0, 0, clientW, clientH}, bgBrush)
	fill(dc, sr(0, 0, 280, h), sideBrush)
	round(dc, sr(40, 42, 112, 114), rgb(8, 145, 178), rgb(34, 211, 238), 16)
	round(dc, sr(28, 206, 252, 292), rgb(255, 255, 255), rgb(220, 229, 238), 12)
	round(dc, sr(292, 28, right, h-48), rgb(255, 255, 255), rgb(220, 229, 238), 12)
	round(dc, sr(324, 212, right-22, 266), rgb(247, 250, 252), rgb(220, 229, 238), 10)
	round(dc, sr(324, 412, right-22, logBottom), rgb(15, 23, 42), rgb(30, 41, 59), 10)

	drawText(dc, "Ready", sr(right-106, 66, right-34, 92), defaultFont, rgb(21, 128, 61), 0x00000004|0x00000020)
}

func fill(dc uintptr, r rect, brush hbrush) {
	procFillRect.Call(dc, uintptr(unsafe.Pointer(&r)), uintptr(brush))
}

func round(dc uintptr, r rect, fillColor, borderColor uintptr, radius int32) {
	brush := createBrushRaw(fillColor)
	pen, _, _ := procCreatePen.Call(0, 1, borderColor)
	oldBrush, _, _ := procSelectObject.Call(dc, brush)
	oldPen, _, _ := procSelectObject.Call(dc, pen)
	procRoundRect.Call(dc, uintptr(r.left), uintptr(r.top), uintptr(r.right), uintptr(r.bottom), uintptr(scale(radius)), uintptr(scale(radius)))
	procSelectObject.Call(dc, oldBrush)
	procSelectObject.Call(dc, oldPen)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(pen)
}

func drawText(dc uintptr, text string, r rect, font hfont, color uintptr, format uint32) {
	oldFont, _, _ := procSelectObject.Call(dc, uintptr(font))
	procSetBkMode.Call(dc, 1)
	procSetTextColor.Call(dc, color)
	procDrawText.Call(dc, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))), ^uintptr(0), uintptr(unsafe.Pointer(&r)), uintptr(format))
	procSelectObject.Call(dc, oldFont)
}

func createFont(name string, height int32, weight int32) hfont {
	pixelHeight := -scale(height)
	ret, _, _ := procCreateFont.Call(
		uintptr(pixelHeight), 0, 0, 0, uintptr(weight), 0, 0, 0,
		1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
	)
	return hfont(ret)
}

func createStatic(parent hwnd, text string, x, y, w, h int32, font hfont) hwnd {
	return createControl(parent, "STATIC", text, wsChild|wsVisible, 0, x, y, w, h, 0, font)
}

func createButton(parent hwnd, text string, id int, x, y, w, h int32) hwnd {
	return createControl(parent, "BUTTON", text, wsChild|wsVisible|bsOwnerDraw|bsPushButton, 0, x, y, w, h, uintptr(id), defaultFont)
}

func createEdit(parent hwnd, text string, x, y, w, h int32, font hfont) hwnd {
	return createControl(parent, "EDIT", text, wsChild|wsVisible|esMultiline|esAutovScroll|esReadOnly|wsVScroll, wsExClientEdge, x, y, w, h, 0, font)
}

func createControl(parent hwnd, className, text string, style, exStyle uint32, x, y, w, h int32, id uintptr, font hfont) hwnd {
	child, _, _ := procCreateWindow.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(style),
		uintptr(scale(x)), uintptr(scale(y)), uintptr(scale(w)), uintptr(scale(h)),
		uintptr(parent), id, 0, 0,
	)
	if font != 0 {
		procSendMessage.Call(child, wmSetFont, uintptr(font), 1)
	}
	return hwnd(child)
}

func move(control hwnd, x, y, w, h int32) {
	if control == 0 {
		return
	}
	procMoveWindow.Call(uintptr(control), uintptr(scale(x)), uintptr(scale(y)), uintptr(scale(w)), uintptr(scale(h)), 1)
}

func clientSize(window hwnd) (int32, int32) {
	var r rect
	procGetClientRect.Call(uintptr(window), uintptr(unsafe.Pointer(&r)))
	return r.right - r.left, r.bottom - r.top
}

func adjustedWindowWidth(clientWidth int32) int32 {
	r := adjustedWindowRect(clientWidth, 760)
	return r.right - r.left
}

func adjustedWindowHeight(clientHeight int32) int32 {
	r := adjustedWindowRect(1120, clientHeight)
	return r.bottom - r.top
}

func adjustedWindowRect(clientWidth, clientHeight int32) rect {
	r := rect{0, 0, scale(clientWidth), scale(clientHeight)}
	procAdjustRect.Call(uintptr(unsafe.Pointer(&r)), launcherWindowStyle, 0, 0)
	return r
}

func defWindowProc(h hwnd, message uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procDefWindowProc.Call(uintptr(h), uintptr(message), wParam, lParam)
	return ret
}

func isMainArea(control hwnd) bool {
	switch control {
	case titleMain, descMain, addressLabel, urlStatic, logTitle, footerStatic:
		return true
	default:
		return false
	}
}

func drawButton(item *drawItemStruct) {
	if item == nil {
		return
	}
	primary := item.ctlID == idOpenDocs
	selected := item.itemState&odsSelected != 0
	fillColor := rgb(248, 250, 252)
	borderColor := rgb(203, 213, 225)
	textColor := rgb(23, 32, 51)
	if primary {
		fillColor = rgb(8, 145, 178)
		borderColor = rgb(8, 145, 178)
		textColor = rgb(255, 255, 255)
	}
	if selected {
		if primary {
			fillColor = rgb(14, 116, 144)
			borderColor = rgb(14, 116, 144)
		} else {
			fillColor = rgb(226, 232, 240)
		}
	}

	r := item.rcItem
	round(uintptr(item.hdc), r, fillColor, borderColor, 8)
	text := buttonText(item.ctlID)
	drawText(uintptr(item.hdc), text, r, defaultFont, textColor, 0x00000001|0x00000004|0x00000020)
}

func buttonText(id uint32) string {
	switch id {
	case idOpenDocs:
		return "打开文档"
	case idOpenFolder:
		return "打开 docs 文件夹"
	case idCopyURL:
		return "复制访问地址"
	default:
		return ""
	}
}

func createSolidBrush(r, g, b byte) hbrush {
	return hbrush(createBrushRaw(rgb(r, g, b)))
}

func createBrushRaw(color uintptr) uintptr {
	brush, _, _ := procCreateBrush.Call(color)
	return brush
}

func rgb(r, g, b byte) uintptr {
	return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16
}

func enableDPIAwareness() {
	// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4.
	ret, _, _ := procSetDpiCtx.Call(^uintptr(3))
	if ret == 0 {
		procSetDpiAware.Call()
	}
}

func updateDPI(window hwnd) {
	if procGetDpiForWin.Find() == nil && window != 0 {
		if dpi, _, _ := procGetDpiForWin.Call(uintptr(window)); dpi != 0 {
			uiDPI = int32(dpi)
			return
		}
	}
	uiDPI = 96
}

func scale(v int32) int32 {
	return v * uiDPI / 96
}

func unscale(v int32) int32 {
	return v * 96 / uiDPI
}

func sr(left, top, right, bottom int32) rect {
	return rect{scale(left), scale(top), scale(right), scale(bottom)}
}

func loadAppIcon(baseDir string) hicon {
	iconPath := filepath.Join(baseDir, "favicon.ico")
	icon, _, _ := procExtractIcon.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(iconPath))), 0)
	if icon != 0 && icon != 1 {
		return hicon(icon)
	}

	exePath, _ := osExecutable()
	if exePath != "" {
		icon, _, _ = procExtractIcon.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(exePath))), 0)
		if icon != 0 && icon != 1 {
			return hicon(icon)
		}
	}
	icon, _, _ = procLoadIcon.Call(0, uintptr(32512))
	return hicon(icon)
}

func osExecutable() (string, error) {
	var buf [260]uint16
	n, _, err := kernel32.NewProc("GetModuleFileNameW").Call(0, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return "", err
	}
	return syscall.UTF16ToString(buf[:n]), nil
}

func appendLog(line string) {
	if mainWindow == 0 {
		return
	}
	logMu.Lock()
	pendingLogs = append(pendingLogs, line)
	logMu.Unlock()
	procPostMessage.Call(uintptr(mainWindow), wmAppLog, 0, 0)
}

func flushLogs() {
	if logEdit == 0 {
		return
	}
	logMu.Lock()
	lines := append([]string(nil), pendingLogs...)
	pendingLogs = pendingLogs[:0]
	logMu.Unlock()
	for _, line := range lines {
		text := strings.TrimRight(line, "\r\n")
		if text == "" {
			continue
		}
		if textLen(logEdit) > 0 {
			text = "\r\n" + text
		}
		procSendMessage.Call(uintptr(logEdit), emSetSel, ^uintptr(0), ^uintptr(0))
		procSendMessage.Call(uintptr(logEdit), emReplaceSel, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
		procSendMessage.Call(uintptr(logEdit), emScrollCaret, 0, 0)
		procInvalidate.Call(uintptr(logEdit), 0, 1)
	}
}

func textLen(control hwnd) uintptr {
	ret, _, _ := procSendMessage.Call(uintptr(control), wmGetTextLen, 0, 0)
	return ret
}

func setStatus(text string) {
	if statusStatic == 0 {
		return
	}
	procSetWindowText.Call(uintptr(statusStatic), uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
}

func requestStop() {
	stopOnce.Do(func() {
		if appStop != nil {
			go appStop()
		}
	})
}

func openFolder(path string) {
	cmd := exec.Command("explorer", path)
	_ = cmd.Start()
}

func copyURLToClipboard(url string) {
	cmd := exec.Command("cmd", "/c", "echo "+escapeCmd(url)+"| clip")
	_ = cmd.Run()
}

func escapeCmd(s string) string {
	return strings.NewReplacer("&", "^&", "|", "^|", "<", "^<", ">", "^>", "^", "^^").Replace(s)
}
