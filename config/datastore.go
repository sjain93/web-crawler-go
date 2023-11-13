package config

type MemoryStore map[string]interface{}

func GetInMemoryStore() MemoryStore {
	return MemoryStore{}
}
