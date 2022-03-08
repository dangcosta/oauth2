//tutorial done based on article: https://medium.com/@cyantarek/build-your-own-oauth2-server-in-go-7d0f660732c3
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gopkg.in/oauth2.v3/models"

	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

func main() {

	//Manager
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	//Client Store
	clientStore := store.NewClientStore()
	manager.MapClientStorage(clientStore)

	//Auth and error handling
	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error: ", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error: ", re.Error.Error())
	})

	// To get token access http://localhost:8080/token?grant_type=client_credentials&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&scope=all
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		srv.HandleTokenRequest(w, r)
	})

	http.HandleFunc("/credentials", func(w http.ResponseWriter, r *http.Request) {
		//Random string for client id
		clientId := uuid.New().String()[:8]

		//Random string for client secret
		clientSecret := uuid.New().String()[:8]

		// We save the cliend id and the client secret to client store. Here it was used memory store but
		// it could have used redis, mongodb, postgres, etc.
		err := clientStore.Set(clientId, &models.Client{
			ID:     clientId,
			Secret: clientSecret,
			Domain: "http://localhost:8084",
		})

		if err != nil {
			fmt.Println((err.Error()))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode((map[string]string{
			"CLIENT_ID":     clientId,
			"CLIENT_SECRET": clientSecret}))
	})

	// To access the protected part use http://localhost:8080/test?access_token=YOUR_ACCESS_TOKEN
	http.HandleFunc("/protected", validateToken(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, I'm protected"))
	}, srv))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Check if a valid token is given and take action based on that. Take as param: func and server
func validateToken(f http.HandlerFunc, srv *server.Server) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		_, err := srv.ValidationBearerToken(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		f.ServeHTTP(w, r)
	})
}
