package port

import (
	"be-ayaka/internal/core/entity/genshin"
	"context"
)

type CharactersRepository interface {
	UpsertCharacter(ctx context.Context, data *genshin.GiCharacter) (*genshin.GiCharacter, error)
	FetchCharacters(ctx context.Context) ([]*genshin.GiCharacter, int, error)
	GetCharacterByID(ctx context.Context, id string) (*genshin.GiCharacter, error)
}