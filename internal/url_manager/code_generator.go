package urlmanager

import (
	"math/rand"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type CodeGenerator struct {
}

func (gen *CodeGenerator) Gen() string {
	ulid := ulid.MustNew(ulid.Now(), rand.New(rand.NewSource(time.Now().UnixNano())))
	shortCode := strings.ToLower(ulid.String())[:8]
	return shortCode
}
