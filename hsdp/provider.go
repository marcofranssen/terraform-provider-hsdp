package hsdp

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns an instance of the HSDP provider
func Provider(build string) terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"region": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"environment"},
				Description:  descriptions["region"],
			},
			"environment": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"region"},
				Description:  descriptions["environment"],
			},
			"iam_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["iam_url"],
			},
			"idm_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["idm_url"],
			},
			"credentials_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["credentials_url"],
			},
			"oauth2_client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["oauth2_client_id"],
			},
			"oauth2_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: descriptions["oauth2_password"],
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["org_id"],
			},
			"org_admin_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["org_admin_username"],
			},
			"org_admin_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: descriptions["org_admin_password"],
			},
			"shared_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				Description: descriptions["shared_key"],
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: descriptions["secret_key"],
			},
			"cartel_host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["cartel_host"],
			},
			"cartel_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: descriptions["cartel_token"],
			},
			"cartel_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: descriptions["cartel_secret"],
			},
			"cartel_no_tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["cartel_no_tls"],
			},
			"cartel_skip_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: descriptions["cartel_skip_verify"],
			},
			"retry_max": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: descriptions["retry_max"],
			},
			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: descriptions["debug"],
			},
			"debug_log": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["debug_log"],
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hsdp_iam_org":            resourceIAMOrg(),
			"hsdp_iam_group":          resourceIAMGroup(),
			"hsdp_iam_permission":     resourceIAMPermission(),
			"hsdp_iam_role":           resourceIAMRole(),
			"hsdp_iam_proposition":    resourceIAMProposition(),
			"hsdp_iam_application":    resourceIAMApplication(),
			"hsdp_iam_user":           resourceIAMUser(),
			"hsdp_iam_client":         resourceIAMClient(),
			"hsdp_iam_service":        resourceIAMService(),
			"hsdp_iam_mfa_policy":     resourceIAMMFAPolicy(),
			"hsdp_credentials_policy": resourceCredentialsPolicy(),
			"hsdp_container_host":     resourceContainerHost(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hsdp_iam_introspect":     dataSourceIAMIntrospect(),
			"hsdp_iam_user":           dataSourceUser(),
			"hsdp_iam_permissions":    dataSourceIAMPermissions(),
			"hsdp_credentials_access": dataSourceS3CredentialsAccess(),
			"hsdp_credentials_policy": dataSourceCredentialsPolicy(),
		},
		ConfigureFunc: providerConfigure,
	}
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"region":             "The HSDP region to configure for",
		"environment":        "The HSDP environment to configure for",
		"iam_url":            "The HSDP IAM instance URL",
		"idm_url":            "The HSDP IDM instance URL",
		"credentials_url":    "The HSDP S3 Credentials instance URL",
		"oauth2_client_id":   "The OAuth2 client id",
		"oauth2_password":    "The OAuth2 password",
		"org_id":             "The (top level) Organization ID - UUID",
		"org_admin_username": "The username of the Organization Admin",
		"org_admin_password": "The password of the Organization Admin",
		"shared_key":         "The shared key",
		"secret_key":         "The secret key",
		"debug":              "Enable debugging output",
		"debug_log":          "The log file to write debugging output to",
		"cartel_host":        "The Cartel host",
		"cartel_token":       "The Cartel token key",
		"cartel_secret":      "The Cartel secret key",
		"cartel_no_tls":      "Disable TLS for Cartel",
		"cartel_skip_verify": "Skip certificate verificsation",
		"retry_max":          "Maximum number of retries for API requests",
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{}

	config.Region = d.Get("region").(string)
	config.Environment = d.Get("environment").(string)
	config.IAMURL = d.Get("iam_url").(string)
	config.IDMURL = d.Get("idm_url").(string)
	config.OAuth2ClientID = d.Get("oauth2_client_id").(string)
	config.OAuth2Secret = d.Get("oauth2_password").(string)
	config.RootOrgID = d.Get("org_id").(string)
	config.OrgAdminUsername = d.Get("org_admin_username").(string)
	config.OrgAdminPassword = d.Get("org_admin_password").(string)
	config.SharedKey = d.Get("shared_key").(string)
	config.SecretKey = d.Get("secret_key").(string)
	config.Debug = d.Get("debug").(bool)
	config.DebugLog = d.Get("debug_log").(string)
	config.S3CredsURL = d.Get("credentials_url").(string)
	config.CartelHost = d.Get("cartel_host").(string)
	config.CartelToken = d.Get("cartel_token").(string)
	config.CartelSecret = d.Get("cartel_secret").(string)
	config.CartelNoTLS = d.Get("cartel_no_tls").(bool)
	config.CartelSkipVerify = d.Get("cartel_skip_verify").(bool)
	config.RetryMax = d.Get("retry_max").(int)

	config.setupIAMClient()
	config.setupS3CredsClient()
	config.setupCartelClient()

	return config, nil
}
