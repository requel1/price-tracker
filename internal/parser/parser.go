package parser

import (
	"context"
	"errors"

	"github.com/requel1/price-tracker/internal/model"
)

var ErrNotSupported = errors.New("site not supported")

type Parser interface {
	Parse(ctx context.Context, url string) (*model.Price, error)
}

func GetParser(site string) (Parser, error) {
	switch site {
	case "chitai-gorod":
		return NewChitaiGorodParser(), nil
	default:
		return nil, ErrNotSupported
	}
}
