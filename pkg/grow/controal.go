package grow

// ControlSignal 控制信号，用于通知 dispatch 循环输入已结束。
type ControlSignal struct {
	Source string
}
