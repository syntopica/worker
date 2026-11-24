package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func recordJudgementsBatch(ctx context.Context, tx store.Conn, now float64, limit int) (int, error) {
	rows, err := queryMaps(ctx, tx, "judges_to_record", params{keyLimit: limit})
	if err != nil {
		return 0, err
	}
	for _, row := range rows {
		if err := recordJudgements(ctx, tx, row, now); err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}

func recordJudgements(ctx context.Context, tx store.Conn, row map[string]any, now float64) error {
	input, err := loadJobInput(ctx, tx, text(row[keyID]))
	if err != nil {
		return err
	}
	output, err := decodeValue([]byte(text(row[keyBody])))
	if err != nil {
		return errInvalidOutput
	}
	verdict := object(object(output)["json"])
	scores := object(verdict["scores"])
	for letter, raw := range object(input[keyLabels]) {
		executor := object(raw)
		score := sqlValue(scores[letter])
		if score == nil {
			score = 0
		}
		err := run(ctx, tx, "record_judgement", params{
			keyGroup: row["shadow_of"], keyQueue: row[keyQueue], keyNow: now,
			keyProvider: knownLabel(executor[keyProvider]), keyModel: knownLabel(executor[keyModel]),
			"score": score, keyBest: verdict[keyBest] == letter,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func knownLabel(value any) string {
	if name := text(value); name != "" {
		return name
	}
	return "unknown"
}
