package grow

// ControlSignal 控制信号，用于通知 sprout 调度器种子入口已关闭。
type ControlSignal struct {
	Source string
}
