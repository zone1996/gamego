定点执行：时、分、秒

固定频率：首次执行时间，间隔固定时间执行一次

ScheduleAtFixedTime

ScheduleAtFixedRate


type TimerTask struct {
	Name String
	
	cancelled bool
	period time.Duration
	turn int
	slot int 
}




