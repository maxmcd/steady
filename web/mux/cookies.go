package mux

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

func setToken(w http.ResponseWriter, token string) {
	c := &http.Cookie{Name: "token", Value: token}
	http.SetCookie(w, c)
}
func setFlash(w http.ResponseWriter, msgs []string) {
	c := &http.Cookie{Name: "flash", Value: encodeFlash(msgs)}
	http.SetCookie(w, c)
}

func getFlash(w http.ResponseWriter, r *http.Request) []string {
	c, err := r.Cookie("flash")
	if err != nil {
		return nil
	}
	value, err := decodeFlash(c.Value)
	if err != nil {
		return nil
	}
	// Delete cookie
	dc := &http.Cookie{Name: "flash", MaxAge: -1, Expires: time.Unix(1, 0)}
	http.SetCookie(w, dc)
	return value
}

func getToken(w http.ResponseWriter, r *http.Request) string {
	// Best effort
	c, err := r.Cookie("token")
	if err != nil {
		return ""
	}
	return c.Value
}

func deleteToken(w http.ResponseWriter, r *http.Request) {
	dc := &http.Cookie{Name: "flash", MaxAge: -1, Expires: time.Unix(1, 0)}
	http.SetCookie(w, dc)
}

func encodeFlash(src []string) string {
	v, _ := json.Marshal(src)
	return base64.RawURLEncoding.EncodeToString(v)
}

func decodeFlash(src string) (v []string, err error) {
	b, err := base64.RawURLEncoding.DecodeString(src)
	if err != nil {
		return nil, err
	}
	return v, json.Unmarshal(b, &v)
}
