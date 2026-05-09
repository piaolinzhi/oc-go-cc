package router

import (
	"fmt"

	"oc-go-cc/internal/complexity"
	"oc-go-cc/internal/config"
)

type ModelRouter struct {
	atomic *config.AtomicConfig
}

func NewModelRouter(atomic *config.AtomicConfig) *ModelRouter {
	return &ModelRouter{atomic: atomic}
}

type RouteResult struct {
	Primary   config.ModelConfig
	Fallbacks []config.ModelConfig
	Scenario  Scenario
	Reason    string
}

func (r *ModelRouter) Route(messages []MessageContent, tokenCount int, modelID string) (RouteResult, error) {
	cfg := r.atomic.Get()
	result := DetectScenario(messages, tokenCount, cfg, modelID)

	primary, ok := cfg.Models[string(result.Scenario)]
	if !ok {
		primary, ok = cfg.Models["default"]
		if !ok {
			return RouteResult{}, fmt.Errorf("no default model configured")
		}
	}

	fallbacks := cfg.Fallbacks[string(result.Scenario)]
	if len(fallbacks) == 0 {
		fallbacks = cfg.Fallbacks["default"]
	}

	return RouteResult{
		Primary:   primary,
		Fallbacks: fallbacks,
		Scenario:  result.Scenario,
		Reason:    result.Reason,
	}, nil
}

func (r *ModelRouter) RouteWithComplexity(compReq *complexity.Request, modelID string) (RouteResult, error) {
	cfg := r.atomic.Get()
	result := DetectScenarioWithComplexity(compReq, cfg, modelID)

	primary, ok := cfg.Models[string(result.Scenario)]
	if !ok {
		primary, ok = cfg.Models["default"]
		if !ok {
			return RouteResult{}, fmt.Errorf("no default model configured")
		}
	}

	fallbacks := cfg.Fallbacks[string(result.Scenario)]
	if len(fallbacks) == 0 {
		fallbacks = cfg.Fallbacks["default"]
	}

	return RouteResult{
		Primary:   primary,
		Fallbacks: fallbacks,
		Scenario:  result.Scenario,
		Reason:    result.Reason,
	}, nil
}

func (r *ModelRouter) IsStreamingScenarioRoutingEnabled() bool {
	return r.atomic.Get().EnableStreamingScenarioRouting
}

func (rr *RouteResult) GetModelChain() []config.ModelConfig {
	chain := []config.ModelConfig{rr.Primary}
	chain = append(chain, rr.Fallbacks...)
	return chain
}

func (r *ModelRouter) RouteForStreaming(messages []MessageContent, tokenCount int) RouteResult {
	cfg := r.atomic.Get()
	result := RouteForStreaming(messages, tokenCount, cfg)

	primary, ok := cfg.Models[string(result.Scenario)]
	if !ok {
		primary, ok = cfg.Models["fast"]
		if !ok {
			primary = cfg.Models["default"]
		}
	}

	fallbacks := cfg.Fallbacks[string(result.Scenario)]
	if len(fallbacks) == 0 {
		fallbacks = cfg.Fallbacks["fast"]
	}

	return RouteResult{
		Primary:   primary,
		Fallbacks: fallbacks,
		Scenario:  result.Scenario,
		Reason:    result.Reason,
	}
}

func (r *ModelRouter) RouteForStreamingWithComplexity(compReq *complexity.Request) RouteResult {
	cfg := r.atomic.Get()
	result := RouteForStreamingWithComplexity(compReq, cfg)

	primary, ok := cfg.Models[string(result.Scenario)]
	if !ok {
		primary, ok = cfg.Models["fast"]
		if !ok {
			primary = cfg.Models["default"]
		}
	}

	fallbacks := cfg.Fallbacks[string(result.Scenario)]
	if len(fallbacks) == 0 {
		fallbacks = cfg.Fallbacks["fast"]
	}

	return RouteResult{
		Primary:   primary,
		Fallbacks: fallbacks,
		Scenario:  result.Scenario,
		Reason:    result.Reason,
	}
}
