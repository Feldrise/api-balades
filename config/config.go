package config

import (
	"os"

	"feldrise.com/balade/database"
	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/notifications/email"
	"feldrise.com/balade/pkg/payments"
	"feldrise.com/balade/pkg/payments/stripe"
	"feldrise.com/balade/pkg/security"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func doesEnvExists(name string) bool {
	_, exists := os.LookupEnv(name)
	return exists
}

type Constants struct {
	// Constants
	Port           string `yaml:"port"`
	JWTSecret      string `yaml:"jwtSecret"`
	DataPath       string `yaml:"dataPath"`
	BaseURL        string `yaml:"baseURL"`
	ApplicationURL string `yaml:"applicationURL"`

	// OpenAI
	OpenAI struct {
		APIKey string `yaml:"apiKey"`
	} `yaml:"openai"`

	// Email
	EmailCredentials struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"emailCredentials"`

	// Database
	ConnectionString string `yaml:"connectionString"`
}

type Config struct {
	Constants

	// Repositories
	RambleRepository                  dbmodel.RambleRepository
	RambleRegistrationRepository      dbmodel.RambleRegistrationRepository
	RambleRegistrationGroupRepository dbmodel.RambleRegistrationGroupRepository
	GuideRepository                   dbmodel.GuideRepository
	UserPermissionOverrideRepository  dbmodel.UserPermissionOverrideRepository
	UserRepository                    dbmodel.UserRepository
	RoleRepository                    dbmodel.RoleRepository
	PaymentRepository                 dbmodel.PaymentRepository

	// Services
	EmailService   email.EmailService
	PaymentService payments.PaymentService
}

func initViper(configName string) (Constants, error) {
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName(configName)

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		return Constants{}, err
	}

	// At this point, the only error would be a missing config file
	if err != nil {
		err = initViperEnv()

		if err != nil {
			return Constants{}, err
		}
	}

	var constants Constants
	err = viper.Unmarshal(&constants)

	return constants, err
}

func initViperEnv() error {
	if !doesEnvExists("PORT") ||
		!doesEnvExists("JWT_SECRET") ||
		!doesEnvExists("DATA_PATH") ||
		!doesEnvExists("BASE_URL") ||
		!doesEnvExists("APPLICATION_URL") ||
		!doesEnvExists("EMAIL_HOST") ||
		!doesEnvExists("EMAIL_PORT") ||
		!doesEnvExists("EMAIL_EMAIL") ||
		!doesEnvExists("EMAIL_PASSWORD") ||
		!doesEnvExists("CONNECTION_STRING") {
		return &MissingEnvVariableError{}
	}

	viper.SetDefault("port", os.Getenv("PORT"))
	viper.SetDefault("jwtSecret", os.Getenv("JWT_SECRET"))
	viper.SetDefault("dataPath", os.Getenv("DATA_PATH"))
	viper.SetDefault("baseURL", os.Getenv("BASE_URL"))
	viper.SetDefault("applicationURL", os.Getenv("APPLICATION_URL"))
	viper.SetDefault("emailCredentials.host", os.Getenv("EMAIL_HOST"))
	viper.SetDefault("emailCredentials.port", os.Getenv("EMAIL_PORT"))
	viper.SetDefault("emailCredentials.email", os.Getenv("EMAIL_EMAIL"))
	viper.SetDefault("emailCredentials.password", os.Getenv("EMAIL_PASSWORD"))
	viper.SetDefault("connectionString", os.Getenv("CONNECTION_STRING"))

	return nil
}

func New() (*Config, error) {
	config := Config{}

	constants, err := initViper("config")

	if err != nil {
		return nil, err
	}

	config.Constants = constants

	// Database
	databaseSession, err := gorm.Open(postgres.Open(config.ConnectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	database.Migrate(databaseSession)

	config.RambleRepository = dbmodel.NewRambleRepository(databaseSession)
	config.RambleRegistrationRepository = dbmodel.NewRambleRegistrationRepository(databaseSession)
	config.RambleRegistrationGroupRepository = dbmodel.NewRambleRegistrationGroupRepository(databaseSession)
	config.GuideRepository = dbmodel.NewGuideRepository(databaseSession)
	config.UserPermissionOverrideRepository = dbmodel.NewUserPermissionOverrideRepository(databaseSession)
	config.UserRepository = dbmodel.NewUserRepository(databaseSession)
	config.RoleRepository = dbmodel.NewRoleRepository(databaseSession)
	config.PaymentRepository = dbmodel.NewPaymentRepository(databaseSession)

	// Email
	config.EmailService = email.NewEmailService(
		constants.DataPath+"/emails",
		email.Credentials{
			Host:     constants.EmailCredentials.Host,
			Port:     constants.EmailCredentials.Port,
			Email:    constants.EmailCredentials.Email,
			Password: constants.EmailCredentials.Password,
		},
	)

	// Payment Service
	stripeService := stripe.NewStripeService()
	encryptionService := security.NewEncryptionService(constants.JWTSecret)
	config.PaymentService = payments.NewPaymentService(
		config.PaymentRepository,
		config.RambleRegistrationRepository,
		config.RambleRegistrationGroupRepository,
		config.RambleRepository,
		config.GuideRepository,
		stripeService,
		encryptionService,
	)

	return &config, nil
}
