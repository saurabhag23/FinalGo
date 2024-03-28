
package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	
	"net/http"
	
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type APIKeyDetails struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}

type APIKey struct {
	Key    string `json:"key"`
	Expiry int64  `json:"expiry"`
}

var (
	apiKeys       map[string]APIKeyDetails
	apiKeysLock   sync.RWMutex
	expirySeconds int64 = 3600 // API key expiry time in seconds (1 hour)
)

func init() {
	apiKeys = make(map[string]APIKeyDetails)
}

func generateAPIKey(owner string) (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	hasher.Write(key)
	apiKey := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return apiKey, nil
}

func generateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	apiKey, err := generateAPIKey(owner)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error generating API key: %v", err)
		return
	}
	apiKeysLock.Lock()
	defer apiKeysLock.Unlock()
	apiKeys[apiKey] = APIKeyDetails{apiKey, owner, time.Now()}
	fmt.Fprintf(w, "Generated API key for %s: %s", owner, apiKey)
}

func authenticateAPIKey(apiKey string) bool {
	apiKeysLock.RLock()
	defer apiKeysLock.RUnlock()
	keyDetails, ok := apiKeys[apiKey]
	if !ok {
		return false
	}
	if float64(time.Since(keyDetails.CreatedAt).Seconds()) > float64(expirySeconds) {
		delete(apiKeys, apiKey)
		return false
	}
	return true
}

func authenticateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if authenticateAPIKey(apiKey) {
		fmt.Fprintf(w, "API Key authenticated")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized")
	}
}

func listAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	apiKeysLock.RLock()
	defer apiKeysLock.RUnlock()
	keys := make([]APIKey, 0)
	for key, details := range apiKeys {
		if float64(time.Since(details.CreatedAt).Seconds()) > float64(expirySeconds) {
			delete(apiKeys, key)
			continue
		}
		keys = append(keys, APIKey{details.Key, details.CreatedAt.Add(time.Second * time.Duration(expirySeconds)).Unix()})
	}
	jsonBytes, err := json.Marshal(keys)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error listing API keys: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
