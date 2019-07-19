package handler

import "net/http"

func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			username := r.Form.Get("username")
			token := r.Form.Get("token")

			if len(username) < 3 || !IsTokenValid(token) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("用户名错误或token失效"))
				return
			}
			h(w, r)
		})

}
