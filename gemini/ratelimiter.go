package gemini

import (
	"context"
	"math"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	rpm *rate.Limiter
	rpd *rate.Limiter
	tpm *rate.Limiter
}

type RateLimiterOptions struct {
	RPM rate.Limit // Requests per minute
	RPD rate.Limit // Requests per day
	TPM rate.Limit // Tokens per minute

	// Optional bursts to override defaults. If zero or negative, defaults are used:
	//   BurstRPM: ceil(RPM)
	//   BurstRPD: ceil(RPD)
	//   BurstTPM: ceil(TPM)
	BurstRPM int
	BurstRPD int
	BurstTPM int
}

func NewRateLimiter(options RateLimiterOptions) *RateLimiter {
	limiter := &RateLimiter{}

	// RPM: per-minute to per-second; burst equals total per-minute allowance
	if options.RPM == rate.Inf {
		limiter.rpm = rate.NewLimiter(rate.Inf, math.MaxInt/2)
	} else {
		perSecond := rate.Limit(options.RPM / 60)
		burst := int(math.Ceil(float64(options.RPM)))
		if options.BurstRPM > 0 { burst = options.BurstRPM }
		if burst < 1 { burst = 1 }
		limiter.rpm = rate.NewLimiter(perSecond, burst)
	}

	// RPD: convert to per-second with a daily burst equal to total daily allowance
	if options.RPD == rate.Inf {
		limiter.rpd = rate.NewLimiter(rate.Inf, math.MaxInt/2)
	} else {
		perSecond := rate.Limit(options.RPD / 86400)
		burst := int(math.Ceil(float64(options.RPD)))
		if options.BurstRPD > 0 { burst = options.BurstRPD }
		if burst < 1 { burst = 1 }
		limiter.rpd = rate.NewLimiter(perSecond, burst)
	}

	// TPM: convert to per-second with burst equal to total tokens per minute
	if options.TPM == rate.Inf {
		limiter.tpm = rate.NewLimiter(rate.Inf, math.MaxInt/2)
	} else {
		perSecond := rate.Limit(options.TPM / 60)
		burst := int(math.Ceil(float64(options.TPM)))
		if options.BurstTPM > 0 { burst = options.BurstTPM }
		if burst < 1 { burst = 1 }
		limiter.tpm = rate.NewLimiter(perSecond, burst)
	}
	return limiter
}

func (r *RateLimiter) Wait(ctx context.Context, maybeTokens ...int) error {
	// Always enforce RPM and RPD
	if err := r.rpm.Wait(ctx); err != nil {
		return err
	}
	if err := r.rpd.Wait(ctx); err != nil {
		return err
	}
	// If a token budget is specified, also enforce TPM
	if len(maybeTokens) > 0 {
		if err := r.tpm.WaitN(ctx, maybeTokens[0]); err != nil {
			return err
		}
	}
	return nil
}
