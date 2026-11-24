package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

const maxWallDays = 40

func runnerWalls(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	if r.id != r.principal.Name {
		return 0, nil, forbidden()
	}
	b, err := decodeBody(r.raw, false)
	if err != nil {
		return 0, nil, err
	}
	walls, err := wallResets(b, r.now+maxWallDays*daySeconds)
	if err != nil {
		return 0, nil, err
	}
	err = jobs.ExtendRunnerWalls(ctx, r.store, walls, r.now)
	return http.StatusOK, map[string]bool{"ok": true}, err
}

func wallResets(b body, latest float64) (map[string]float64, error) {
	walls := make(map[string]float64, len(b))
	for name, value := range b {
		switch value.(type) {
		case json.Number, bool:
		default:
			return nil, badRequest()
		}
		reset, err := numberValue(value)
		if err != nil || reset > latest {
			return nil, badRequest()
		}
		walls[name] = reset
	}
	return walls, nil
}
