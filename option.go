package cron

// Option represents a modification to the default behavior of a Cron
type Option func(*Cron)

func WithLogger (logger Logger) Option {
	return func (c *Cron){
		c.logger = logger
	}
}

func WithChain(wrappers ...JobWrapper) Option {
	return func(c *Cron) {
		c.chain = NewChain(wrappers...)
	}
}

