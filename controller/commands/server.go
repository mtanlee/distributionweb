package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/mtanlee/distributionweb/auth/builtin"
	"github.com/mtanlee/distributionweb/auth/ldap"
	"github.com/mtanlee/distributionweb/controller/api"
	"github.com/mtanlee/distributionweb/controller/manager"
	"github.com/mtanlee/distributionweb/version"
)

var (
	controllerManager *manager.Manager
)

func CmdServer(c *cli.Context) {
	rethinkdbAddr := c.String("rethinkdb-addr")
	rethinkdbDatabase := c.String("rethinkdb-database")
	rethinkdbAuthKey := c.String("rethinkdb-auth-key")
	disableUsageInfo := c.Bool("disable-usage-info")
	listenAddr := c.String("listen")
	authWhitelist := c.StringSlice("auth-whitelist-cidr")
	enableCors := c.Bool("enable-cors")
	ldapServer := c.String("ldap-server")
	ldapPort := c.Int("ldap-port")
	ldapBaseDn := c.String("ldap-base-dn")
	ldapAutocreateUsers := c.Bool("ldap-autocreate-users")
	ldapDefaultAccessLevel := c.String("ldap-default-access-level")

	log.Infof("distributionweb version %s", version.Version)

	if len(authWhitelist) > 0 {
		log.Infof("whitelisting the following subnets: %v", authWhitelist)
	}

	allowInsecure := c.Bool("allow-insecure")


	// default to builtin auth
	authenticator := builtin.NewAuthenticator("defaultdistributionweb")

	// use ldap auth if specified
	if ldapServer != "" {
		authenticator = ldap.NewAuthenticator(ldapServer, ldapPort, ldapBaseDn, ldapAutocreateUsers, ldapDefaultAccessLevel)
	}

	controllerManager, err := manager.NewManager(rethinkdbAddr, rethinkdbDatabase, rethinkdbAuthKey, disableUsageInfo, authenticator)
	if err != nil {
		log.Fatal(err)
	}


	distributionwebTlsCert := c.String("distributionweb-tls-cert")
	distributionwebTlsKey := c.String("distributionweb-tls-key")
	distributionwebTlsCACert := c.String("distributionweb-tls-ca-cert")

	apiConfig := api.ApiConfig{
		ListenAddr:         listenAddr,
		Manager:            controllerManager,
		AuthWhiteListCIDRs: authWhitelist,
		EnableCORS:         enableCors,
		AllowInsecure:      allowInsecure,
		TLSCACertPath:      distributionwebTlsCACert,
		TLSCertPath:        distributionwebTlsCert,
		TLSKeyPath:         distributionwebTlsKey,
	}

	distributionwebApi, err := api.NewApi(apiConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := distributionwebApi.Run(); err != nil {
		log.Fatal(err)
	}
}
