Golang 实现游戏定时器
刚开始学习golang，发现Go没有游戏里定时器功能，有个cron库实现了golang版的cron功能。
但是里面用了是sort全排序，在游戏里，通常是存在数万个定时器，采用排序算法效率较低，
根据其实现思路，将其改成最小堆实现，同时去掉了一些游戏定时器不需要的内容。
其使用如下：

// 创建定时器
c := cron.New()
// 启动定时器
c.Start()
// 停止定时器
c.Stop()
// 延迟调用函数
c.CallOut(4, f)
c.CallFre(4, f)
c.Daily(12, 5)
c.Weekly(21, 34, 1)
