package eventoutbox

import "time"

func publishTimeout(options PublishOptions) time.Duration {
	if options.PublishTimeout > 0 {
		return options.PublishTimeout
	}
	return defaultPublishTimeout
}

func retryDelay(options PublishOptions, attempt int32) time.Duration {
	if options.RetryDelay != nil {
		return options.RetryDelay(attempt)
	}
	return DefaultRetryDelay(attempt)
}

func optionNow(options PublishOptions) time.Time {
	if options.Now != nil {
		return options.Now()
	}
	return time.Now()
}
