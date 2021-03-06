package app

import (
"context"
"encoding/json"
	 "github.com/DaniilOr/restricted/cmd/service/app/dtos"
	"github.com/DaniilOr/restricted/cmd/service/app/middleware/authenticator"
	"github.com/DaniilOr/restricted/cmd/service/app/middleware/authorizator"
	"github.com/DaniilOr/restricted/cmd/service/app/middleware/identificator"
	"github.com/DaniilOr/restricted/pkg/business"
	"github.com/DaniilOr/restricted/pkg/payments"
	"github.com/DaniilOr/restricted/pkg/security"
	"github.com/go-chi/chi"
"log"
"net/http"
	"strconv"
)
type Server struct {
	securitySvc *security.Service
	businessSvc *business.Service
	paymentsSvc *payments.Service
	router      chi.Router
}

func NewServer(securitySvc *security.Service, businessSvc *business.Service, paymentsSvc *payments.Service, router chi.Router) *Server {
	return &Server{securitySvc: securitySvc, businessSvc: businessSvc, paymentsSvc: paymentsSvc, router: router}
}

func (s *Server) Init() error {
	s.router.Post("/users", s.handleRegister)
	s.router.Put("/users", s.handleLogin)


	identificatorMd := identificator.Identificator
	authenticatorMd := authenticator.Authenticator(
		identificator.Identifier, s.securitySvc.UserDetails,
	)

	// функция-связка между middleware и security service
	// (для чистоты security service ничего не знает об http)
	roleChecker := func(ctx context.Context, roles ...string) bool {
		userDetails, err := authenticator.Authentication(ctx)
		if err != nil {
			return false
		}
		return s.securitySvc.HasAnyRole(ctx, userDetails, roles...)
	}
	adminRoleMd := authorizator.Authorizator(roleChecker, security.RoleAdmin)
	userRoleMd := authorizator.Authorizator(roleChecker, security.RoleUser)

	s.router.Get("/public", s.handlePublic)
	s.router.With(identificatorMd, authenticatorMd, adminRoleMd).Get("/admin", s.handleAdmin)
	s.router.With(identificatorMd, authenticatorMd, userRoleMd).Get("/user", s.handleUser)

	s.router.With(identificatorMd, authenticatorMd, userRoleMd).Get("/user/payments", s.handleGetPayments)
	s.router.With(identificatorMd, authenticatorMd, userRoleMd).Post("/user/payments", s.handlePostPayment)
	s.router.With(identificatorMd, authenticatorMd, adminRoleMd).Get("/admin/payments", s.handleViewPayments)
	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s *Server) handleRegister(writer http.ResponseWriter, request *http.Request) {
	login := request.PostFormValue("login")
	if login == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	password := request.PostFormValue("password")
	if password == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := s.securitySvc.Register(request.Context(), login, password)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data := &dtos.TokenDTO{Token: token}
	respBody, err := json.Marshal(data)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = makeResponse(respBody, http.StatusOK, writer)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleLogin(writer http.ResponseWriter, request *http.Request) {
	login := request.PostFormValue("login")
	if login == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	password := request.PostFormValue("password")
	if password == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := s.securitySvc.Login(request.Context(), login, password)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data := &dtos.TokenDTO{Token: token}
	respBody, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(respBody)
	if err != nil {
		log.Print(err)
	}
}

// Доступно всем
func (s *Server) handlePublic(writer http.ResponseWriter, request *http.Request) {
	_, err := writer.Write([]byte("public"))
	if err != nil {
		log.Print(err)
	}
}

// Только пользователям с ролью ADMIN
func (s *Server) handleAdmin(writer http.ResponseWriter, request *http.Request) {
	_, err := writer.Write([]byte("admin"))
	if err != nil {
		log.Print(err)
	}
}

// Только пользователям с ролью USER
func (s *Server) handleUser(writer http.ResponseWriter, request *http.Request) {
	_, err := writer.Write([]byte("user"))
	if err != nil {
		log.Print(err)
	}
}

func(s*Server) handleGetPayments(writer http.ResponseWriter, request * http.Request){
	id := request.Context().Value(authenticator.AuthenticationContextKey).(*security.UserDetails).ID
	paymentsList, err := s.paymentsSvc.GetUserPayments(request.Context(), id)
	if err != nil{
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	err = makeResponse(paymentsList, http.StatusOK, writer)
	if err != nil{
		log.Println(err)
		return
	}
}

func(s*Server) handlePostPayment(writer http.ResponseWriter, request * http.Request){
	id := request.Context().Value(authenticator.AuthenticationContextKey).(*security.UserDetails).ID
	amountS := request.PostFormValue("amount")
	if amountS == ""{
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseInt(amountS, 10, 64)
	if err != nil{
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	err = s.paymentsSvc.AddUserPayments(request.Context(), id, amount)
	if err != nil{
		log.Println(err)
		cerr := makeResponse(dtos.ResultDTO{Result: "Error"}, http.StatusOK, writer)
		if cerr != nil{
			log.Println(err)
			return
		}
		return
	}
	result := dtos.ResultDTO{Result: "Done"}
	err = makeResponse(result, http.StatusOK, writer)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}
func (s*Server) handleViewPayments(writer http.ResponseWriter, request * http.Request){
	paymentsList, err := s.paymentsSvc.GetAllPayments(request.Context())
	if err != nil{
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = makeResponse(paymentsList, http.StatusOK, writer)
	if err != nil{
		log.Println(err)
		return
	}
}
func makeResponse(resp interface{}, status int, w http.ResponseWriter) error {
	if resp != nil {
		body, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		w.Header().Add("Content-Type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
	}
	w.WriteHeader(status)
	return nil
}

