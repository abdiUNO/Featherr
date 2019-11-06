package auth

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/abdiUNO/featherr/config"
	u "github.com/abdiUNO/featherr/utils"
	"github.com/abdiUNO/featherr/utils/response"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var CreateUser = func(w http.ResponseWriter, r *http.Request) {

	user := &User{}
	err := json.NewDecoder(r.Body).Decode(user) //decode the request body into struct and failed if any error occur
	if err != nil {
		response.HandleError(w, u.NewError(u.EINTERNAL, "Invalid request", err))
		return
	}

	if validErr := user.Validate(); validErr != nil {
		response.HandleError(w, validErr)
		return
	}

	data, ormErr := user.Create()
	if ormErr != nil {
		response.HandleError(w, u.NewError(u.EINTERNAL, "Internal server err", ormErr))
		return
	}

	response.Json(w, map[string]interface{}{
		"user": data,
	})

}

var Authenticate = func(w http.ResponseWriter, r *http.Request) {

	user := &User{}
	//decode the request body into struct and failed if any error occur
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		response.HandleError(w, u.NewError(u.EINTERNAL, "Invalid request", err))
		return
	}

	data, err := Login(user.Email, user.Password, user.FcmToken)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"user": data,
	})

}

var UpdateUser = func(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userId := params["id"]
	user := &User{}
	//decode the request body into struct and failed if any error occur
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		response.HandleError(w, u.NewError(u.EINTERNAL, "Invalid request", err))
		return
	}

	user, err := Update(userId, user)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"user": user,
	})

}

type ChangePasswordBody struct {
	OldPassword string `json:",oldPassword"`
	NewPassword string `json:",newPassword"`
}

var ChangePassword = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*Token)
	user := GetUser(token.UserId)

	jsonBody := &ChangePasswordBody{}
	//decode the request body into struct and failed if any error occur
	if err := json.NewDecoder(r.Body).Decode(jsonBody); err != nil {
		response.HandleError(w, u.NewError(u.EINTERNAL, "Invalid request", err))
		return
	}

	updateErr := user.UpdatePassword(jsonBody.OldPassword, jsonBody.NewPassword)

	if updateErr != nil {
		response.HandleError(w, updateErr)
		return
	}

	response.Json(w, map[string]interface{}{
		"data": "Updated user password",
	})

}

var FindUsers = func(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	users, err := QueryUsers(query)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"users": users,
	})

}

func DeleteTempFiles(userId string) {
	dir, _ := os.Getwd()

	imageName := fmt.Sprintf("%s_%d.jpg", userId, 480)
	imagePath := fmt.Sprintf("%s/tmp/%s", dir, imageName)

	if ok := os.Remove(imagePath); ok != nil {
		log.Println(ok)
	}

	thumbName := fmt.Sprintf("%s_%d.jpg", userId, 200)
	thumbPath := fmt.Sprintf("%s/tmp/%s", dir, thumbName)

	if ok := os.Remove(thumbPath); ok != nil {
		log.Println(ok)
	}

	return
}

var UploadProfileImage = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*Token)
	user := GetUser(token.UserId)

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	imageUrl, err := uploadFile(r, user.ID)

	user.Image = imageUrl

	GetDB().Save(&user)
	if err != nil {
		response.HandleError(w, u.NewError(u.EINTERNAL, "Invalid request", err))
		return
	}

	go DeleteTempFiles(user.ID)

	response.Json(w, map[string]interface{}{
		"fileUrl": imageUrl,
	})
}

func resizeImage(file multipart.File, id string, size int) (string, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		log.Println(err)
		return "", err
	}

	m := resize.Resize(uint(size), 0, img, resize.Lanczos3)

	fileName := fmt.Sprintf("%s_%d.jpg", id, size)
	dir, _ := os.Getwd()
	dirPath := fmt.Sprintf("%s/tmp/%s", dir, fileName)

	out, err := os.Create(dirPath)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer out.Close()

	err = jpeg.Encode(out, m, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return out.Name(), nil
}

func UploadObject(client *storage.Client, name string, contentType string) error {
	file, err := os.Open(name) // Maybe add complete path
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	// Creates a client.

	if err != nil {
		log.Println(err)
		return err
	}

	cfg := config.GetConfig()

	bucket := client.Bucket(cfg.BucketName)

	ctx := context.Background()
	writer := bucket.Object("profile_images/" + filepath.Base(name)).NewWriter(ctx)

	// Warning: storage.AllUsers gives public read access to anyone.
	writer.ContentType = contentType

	// Entries are immutable, be aggressive about caching (1 day).
	writer.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(writer, file); err != nil {
		log.Println(err)
		return err
	}

	if err := writer.Close(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func uploadFile(r *http.Request, userId string) (string, error) {
	f, fh, err := r.FormFile("image")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer f.Close()

	profileImg, fileErr := resizeImage(f, userId, 480)
	if fileErr != nil {
		log.Println(err)
		return "", fileErr
	}

	file, err := os.Open(profileImg) // Maybe add complete path
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer file.Close()

	thumbImg, fileErr := resizeImage(file, userId, 200)
	if fileErr != nil {
		log.Println(err)
		return "", fileErr
	}

	client, err := storage.NewClient(r.Context())
	if err != nil {
		log.Println(err)
		return "", err
	}

	if ok := UploadObject(client, profileImg, fh.Header.Get("Content-Type")); ok != nil {
		log.Println(ok)
		return "", ok
	}

	if ok := UploadObject(client, thumbImg, fh.Header.Get("Content-Type")); ok != nil {
		log.Println(ok)
		return "", ok
	}

	const publicURL = "https://storage.googleapis.com/%s/%s"

	return fmt.Sprintf(publicURL, "featherr/profile_images", userId), nil

}

var GenerateOTP = func(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userId := params["id"]
	user, dbErr := FindUserById(userId)

	if dbErr != nil {
		response.HandleError(w, u.NewError(u.ENOTFOUND, "could not find user", dbErr))
		return
	}

	code, err := CreateCode(user)
	if err != nil {
		response.HandleError(w, u.NewError(u.EINTERNAL, "could not create code", err))
		return
	}

	err = EmailCode(r.Context(), code, user)
	if err != nil {
		fmt.Println(err.Error())
		response.HandleError(w, u.NewError(u.EINTERNAL, "could not send otp email", err))
		return
	}

	response.Json(w, map[string]interface{}{
		"codeSent": true,
	})
}

var ValidateOTP = func(w http.ResponseWriter, r *http.Request) {
	passcode := r.FormValue("code")
	params := mux.Vars(r)
	userId := params["id"]
	user, dbErr := FindUserById(userId)

	if dbErr != nil {
		response.HandleError(w, u.NewError(u.ENOTFOUND, "could not find user", dbErr))
		return
	}

	isValid, err := ValidateCode(passcode, user)

	if err != nil {
		response.HandleError(w, u.NewError(u.EINTERNAL, "could not validate code", err))
		return
	}

	if isValid == true {
		user.EmailVerified = true
		dbErr := GetDB().Save(&user).Error

		if dbErr != nil {
			response.HandleError(w, u.NewError(u.EINTERNAL, "could not update user", nil))
			return
		}
	}

	response.Json(w, map[string]interface{}{
		"isValid": isValid,
	})
}
