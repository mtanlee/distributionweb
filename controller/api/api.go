package api

import (
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mailgun/oxy/forward"
	"github.com/mtanlee/distributionweb/auth"
	"github.com/mtanlee/distributionweb/controller/manager"
	"github.com/mtanlee/distributionweb/controller/middleware/access"
	"github.com/mtanlee/distributionweb/controller/middleware/audit"
	mAuth "github.com/mtanlee/distributionweb/controller/middleware/auth"
	"github.com/mtanlee/distributionweb/tlsutils"
	
)

type (
	Api struct {
		listenAddr         string
		manager            manager.Manager
		authWhitelistCIDRs []string
		enableCors         bool
		serverVersion      string
		allowInsecure      bool
		tlsCACertPath      string
		tlsCertPath        string
		tlsKeyPath         string
		dUrl               string
		fwd                *forward.Forwarder
	}

	ApiConfig struct {
		ListenAddr         string
		Manager            manager.Manager
		AuthWhiteListCIDRs []string
		EnableCORS         bool
		AllowInsecure      bool
		TLSCACertPath      string
		TLSCertPath        string
		TLSKeyPath         string
	}

	Credentials struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}
)

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

func NewApi(config ApiConfig) (*Api, error) {
	return &Api{
		listenAddr:         config.ListenAddr,
		manager:            config.Manager,
		authWhitelistCIDRs: config.AuthWhiteListCIDRs,
		enableCors:         config.EnableCORS,
		allowInsecure:      config.AllowInsecure,
		tlsCertPath:        config.TLSCertPath,
		tlsKeyPath:         config.TLSKeyPath,
		tlsCACertPath:      config.TLSCACertPath,
	}, nil
}

func (a *Api) Run() error {
	globalMux := http.NewServeMux()
	controllerManager := a.manager


	apiRouter := mux.NewRouter()
	apiRouter.HandleFunc("/api/accounts", a.accounts).Methods("GET")
	apiRouter.HandleFunc("/api/accounts", a.saveAccount).Methods("POST")
	apiRouter.HandleFunc("/api/accounts/{username}", a.account).Methods("GET")
	apiRouter.HandleFunc("/api/accounts/{username}", a.deleteAccount).Methods("DELETE")
	apiRouter.HandleFunc("/api/roles", a.roles).Methods("GET")
	apiRouter.HandleFunc("/api/roles/{name}", a.role).Methods("GET")
	apiRouter.HandleFunc("/api/events", a.events).Methods("GET")
	apiRouter.HandleFunc("/api/events", a.purgeEvents).Methods("DELETE")
	apiRouter.HandleFunc("/api/registries", a.registries).Methods("GET")
	apiRouter.HandleFunc("/api/registries", a.addRegistry).Methods("POST")
	apiRouter.HandleFunc("/api/registries/{name}", a.registry).Methods("GET")
	apiRouter.HandleFunc("/api/registries/{name}", a.removeRegistry).Methods("DELETE")
	apiRouter.HandleFunc("/api/registries/{name}/repositories", a.repositories).Methods("GET")
	apiRouter.HandleFunc("/api/registries/{name}/repositories/{repo:.*}", a.repository).Methods("GET")
	apiRouter.HandleFunc("/api/registries/{name}/repositories/{repo:.*}", a.deleteRepository).Methods("DELETE")
	apiRouter.HandleFunc("/api/servicekeys", a.serviceKeys).Methods("GET")
	apiRouter.HandleFunc("/api/servicekeys", a.addServiceKey).Methods("POST")
	apiRouter.HandleFunc("/api/servicekeys", a.removeServiceKey).Methods("DELETE")
	apiRouter.HandleFunc("/api/webhookkeys", a.webhookKeys).Methods("GET")
	apiRouter.HandleFunc("/api/webhookkeys/{id}", a.webhookKey).Methods("GET")
	apiRouter.HandleFunc("/api/webhookkeys", a.addWebhookKey).Methods("POST")
	apiRouter.HandleFunc("/api/webhookkeys/{id}", a.deleteWebhookKey).Methods("DELETE")
	apiRouter.HandleFunc("/api/consolesession/{container}", a.createConsoleSession).Methods("GET")
	apiRouter.HandleFunc("/api/consolesession/{token}", a.consoleSession).Methods("GET")
	apiRouter.HandleFunc("/api/consolesession/{token}", a.removeConsoleSession).Methods("DELETE")

	// global handler
	globalMux.Handle("/", http.FileServer(http.Dir("static")))

	auditExcludes := []string{
		"^/api/events",
	}
	apiAuditor := audit.NewAuditor(controllerManager, auditExcludes)

	// api router; protected by auth
	apiAuthRouter := negroni.New()
	apiAuthRequired := mAuth.NewAuthRequired(controllerManager, a.authWhitelistCIDRs)
	apiAccessRequired := access.NewAccessRequired(controllerManager)
	apiAuthRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	apiAuthRouter.Use(negroni.HandlerFunc(apiAccessRequired.HandlerFuncWithNext))
	apiAuthRouter.Use(negroni.HandlerFunc(apiAuditor.HandlerFuncWithNext))
	apiAuthRouter.UseHandler(apiRouter)
	globalMux.Handle("/api/", apiAuthRouter)

	// account router ; protected by auth
	accountRouter := mux.NewRouter()
	accountRouter.HandleFunc("/account/changepassword", a.changePassword).Methods("POST")
	accountAuthRouter := negroni.New()
	accountAuthRequired := mAuth.NewAuthRequired(controllerManager, a.authWhitelistCIDRs)
	accountAuthRouter.Use(negroni.HandlerFunc(accountAuthRequired.HandlerFuncWithNext))
	accountAuthRouter.Use(negroni.HandlerFunc(apiAuditor.HandlerFuncWithNext))
	accountAuthRouter.UseHandler(accountRouter)
	globalMux.Handle("/account/", accountAuthRouter)

	// login handler; public
	loginRouter := mux.NewRouter()
	loginRouter.HandleFunc("/auth/login", a.login).Methods("POST")
	globalMux.Handle("/auth/", loginRouter)
	
	// hub handler; public
	hubRouter := mux.NewRouter()
	hubRouter.HandleFunc("/hub/webhook/{id}", a.hubWebhook).Methods("POST")
	globalMux.Handle("/hub/", hubRouter)

	
	// check for admin user
	if _, err := controllerManager.Account("admin"); err == manager.ErrAccountDoesNotExist {
		// create roles
		acct := &auth.Account{
			Username:  "admin",
			Password:  "admin",
			FirstName: "mtan",
			LastName:  "lee",
			Roles:     []string{"admin"},
		}
		if err := controllerManager.SaveAccount(acct); err != nil {
			log.Fatal(err)
		}
		log.Infof("created admin user: username: admin password: admin")
	}

	log.Infof("controller listening on %s", a.listenAddr)

	s := &http.Server{
		Addr:    a.listenAddr,
		Handler: context.ClearHandler(globalMux),
	}

	var runErr error

	if a.tlsCertPath != "" && a.tlsKeyPath != "" {
		log.Infof("using TLS for communication: cert=%s key=%s",
			a.tlsCertPath,
			a.tlsKeyPath,
		)

		// setup TLS config
		var caCert []byte
		if a.tlsCACertPath != "" {
			ca, err := ioutil.ReadFile(a.tlsCACertPath)
			if err != nil {
				return err
			}

			caCert = ca
		}

		serverCert, err := ioutil.ReadFile(a.tlsCertPath)
		if err != nil {
			return err
		}

		serverKey, err := ioutil.ReadFile(a.tlsKeyPath)
		if err != nil {
			return err
		}

		tlsConfig, err := tlsutils.GetServerTLSConfig(caCert, serverCert, serverKey, a.allowInsecure)
		if err != nil {
			return err
		}

		s.TLSConfig = tlsConfig

		runErr = s.ListenAndServeTLS(a.tlsCertPath, a.tlsKeyPath)
	} else {
		runErr = s.ListenAndServe()
	}

	return runErr
}
