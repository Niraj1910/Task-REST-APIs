package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("Failed to hash password: " + err.Error())
	}
	return string(hashedBytes)
}

func CompareHashedPassword(hashedPassword, password string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		panic("password is incorrect  " + password + ": " + err.Error())
	}
	return err == nil
}

func UserIDFromContext(ctx *gin.Context) (uint, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return 0, false
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return 0, false
	}
	return uid, true
}

func SendVerificationMail(userName, toEmail, verifyLink string) error {

	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	fromAddr := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	resendApiKey := os.Getenv("RESEND_API_KEY")

	if host == "" || fromAddr == "" || password == "" {
		log.Warn().Msg("SMTP config missing - skipping mail send")
		return nil
	}

	port := 587 // default
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	// load html template
	tmplPath := filepath.Join("template", "verifyEmail.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse verifyEmail.html template")
		return nil
	}

	// template data format
	data := struct {
		Name       string
		VerifyLink string
		Time       string
	}{
		Name:       userName,
		VerifyLink: verifyLink,
		Time:       "10 minutes",
	}

	var htmlBody bytes.Buffer
	err = tmpl.Execute(&htmlBody, data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render email template")
		return err
	}

	// send mail using resend
	client := resend.NewClient(resendApiKey)

	params := &resend.SendEmailRequest{
		From:    fromAddr,
		To:      []string{toEmail},
		Subject: "Golang Task API user registration confirmation",
		Html:    htmlBody.String(),
	}

	_, err = client.Emails.Send(params)
	if err != nil {
		log.Error().Err(err).Str("to", toEmail).Msg("Failed to send verification mail")
		return err
	}

	log.Info().Str("to", toEmail).Msg("Verification mail sent successfully")
	return nil
}

func InitLogger() {

	if os.Getenv("ENV") == "development" || os.Getenv("ENV") == "" {

		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "00:00:00",
			NoColor:    true,
		})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = log.Output(os.Stdout)
	}
}
