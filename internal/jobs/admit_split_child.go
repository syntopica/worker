package jobs

import (
	"context"
	"net/http"
	"regexp"
	"strconv"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/store"
)

const minSplit = 2

var splitSuffix = regexp.MustCompile(`/(0|[1-9]\d*)/([1-9]\d*)$`)

// admitSplitChild checks a split child against its parent and supersedes the
// parent: the child's key is the parent's followed by /<index>/<count>, every
// child names the same count, and children do not count against the
// outstanding limit.
func admitSplitChild(
	ctx context.Context, c store.Conn, cfg *config.Config, producer string, sub *submission, now float64,
) error {
	var parent struct {
		producer, queue, state, key string
		count                       int
	}
	found, err := queryRow(ctx, c, "split_parent", params{keyID: *sub.ParentID},
		&parent.producer, &parent.queue, &parent.state, &parent.count, &parent.key)
	if err != nil {
		return err
	}
	if !found || parent.producer != producer || parent.queue != sub.Queue {
		return bad(codeBadSplitChild)
	}
	count, err := splitCount(sub.Key, parent.key, parent.state, cfg.MaxSplit)
	if err != nil {
		return err
	}
	if parent.count != 0 && parent.count != count {
		return fault.New(http.StatusConflict, "split_count_mismatch")
	}

	return run(ctx, c, "supersede_parent", params{keyID: *sub.ParentID, "count": count, keyNow: now})
}

func splitCount(key, parent, state string, maximum int) (int, error) {
	suffix := splitSuffix.FindStringSubmatch(key)
	if len(suffix) == 0 || key != parent+suffix[0] {
		return 0, bad(codeBadSplitChild)
	}
	if state != stateSplitRequested && state != stateSuperseded {
		return 0, fault.New(http.StatusConflict, "parent_not_splitting")
	}
	index, indexErr := strconv.Atoi(suffix[1])
	count, countErr := strconv.Atoi(suffix[2])
	if indexErr != nil || countErr != nil || count < minSplit || count > maximum || index >= count {
		return 0, bad("bad_split_count")
	}
	return count, nil
}
