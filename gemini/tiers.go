package gemini

import "golang.org/x/time/rate"

var (
	Tier1RateLimits = map[ModelName]RateLimiterOptions{
		Pro: {
			RPM: 150,
			RPD: 2_000_000,
			TPM: 10_000,
		},
		Flash: {
			RPM: 1_000,
			RPD: 1_000_000,
			TPM: 10_000,
		},
		FlashLite: {
			RPM: 4_000,
			RPD: 4_000_000,
			TPM: rate.Inf,
		},
		Embedding001: {
			RPM: 3_000,
			RPD: 10_000_000,
			TPM: rate.Inf,
		},
	}
	Tier2RateLimits = map[ModelName]RateLimiterOptions{
		Pro: {
			RPM: 1_000,
			RPD: 5_000_000,
			TPM: 50_000,
		},
		Flash: {
			RPM: 2_000,
			RPD: 3_000_000,
			TPM: 100_000,
		},
		FlashLite: {
			RPM: 10_000,
			RPD: 10_000_000,
			TPM: rate.Inf,
		},
		Embedding001: {
			RPM: 3_000,
			RPD: 10_000_000,
			TPM: rate.Inf,
		},
	}
	Tier3RateLimits = map[ModelName]RateLimiterOptions{
		Pro: {
			RPM: 2_000,
			RPD: rate.Inf,
			TPM: 8_000_000,
		},
		Flash: {
			RPM: 10_000,
			RPD: rate.Inf,
			TPM: 8_000_000,
		},
		FlashLite: {
			RPM: 30_000,
			RPD: rate.Inf,
			TPM: 30_000_000,
		},
		Embedding001: {
			RPM: 3_000,
			RPD: 10_000_000,
			TPM: rate.Inf,
		},
	}
	TokenCountRateLimiter = NewRateLimiter(RateLimiterOptions{
		RPM: 3000,
		RPD: 10_000_000,
		TPM: rate.Inf,
	})

	Tier1Models = map[ModelName]*Model{
		Pro:          NewModel(Pro, NewRateLimiter(Tier1RateLimits[Pro]), TokenCountRateLimiter),
		Flash:        NewModel(Flash, NewRateLimiter(Tier1RateLimits[Flash]), TokenCountRateLimiter),
		FlashLite:    NewModel(FlashLite, NewRateLimiter(Tier1RateLimits[FlashLite]), TokenCountRateLimiter),
		Embedding001: NewModel(Embedding001, NewRateLimiter(Tier1RateLimits[Embedding001]), TokenCountRateLimiter),
	}
	Tier2Models = map[ModelName]*Model{
		Pro:          NewModel(Pro, NewRateLimiter(Tier2RateLimits[Pro]), TokenCountRateLimiter),
		Flash:        NewModel(Flash, NewRateLimiter(Tier2RateLimits[Flash]), TokenCountRateLimiter),
		FlashLite:    NewModel(FlashLite, NewRateLimiter(Tier2RateLimits[FlashLite]), TokenCountRateLimiter),
		Embedding001: NewModel(Embedding001, NewRateLimiter(Tier2RateLimits[Embedding001]), TokenCountRateLimiter),
	}
	Tier3Models = map[ModelName]*Model{
		Pro:          NewModel(Pro, NewRateLimiter(Tier3RateLimits[Pro]), TokenCountRateLimiter),
		Flash:        NewModel(Flash, NewRateLimiter(Tier3RateLimits[Flash]), TokenCountRateLimiter),
		FlashLite:    NewModel(FlashLite, NewRateLimiter(Tier3RateLimits[FlashLite]), TokenCountRateLimiter),
		Embedding001: NewModel(Embedding001, NewRateLimiter(Tier3RateLimits[Embedding001]), TokenCountRateLimiter),
	}
)
