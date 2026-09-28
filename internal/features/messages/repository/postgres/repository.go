package messages_postgres_repository

import core_postgres_pool "github.com/Rics69/rics-chat/internal/core/repository/postgres/pool"

type MessagesRepository struct {
	pool core_postgres_pool.Pool
}

func NewMessagesRepository(pool core_postgres_pool.Pool) *MessagesRepository {
	return &MessagesRepository{
		pool: pool,
	}
}
