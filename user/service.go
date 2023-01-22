package user

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	RegisterUser(input RegisterUserInput) (User, error)
	Login(input LoginInput) (User, error)
	IsEmailAvailable(input CheckEmailInput) (bool, error)
	SaveAvatar(ID int, fileLocation string) (User, error)
	GetUserByID(ID int) (User, error)
	//UploadToCloud(file *multipart.FileHeader, userId int) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) *service {
	return &service{repository}
}

func (s *service) RegisterUser(input RegisterUserInput) (User, error) {
	user := User{}
	user.Name = input.Name
	user.Email = input.Email
	user.Occupation = input.Occupation
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.MinCost)
	if err != nil {
		return user, err
	}

	user.PasswordHash = string(passwordHash)
	user.Role = "user"

	newUser, err := s.repository.Save(user)
	if err != nil {
		return newUser, err
	}

	return newUser, nil
}

func (s *service) Login(input LoginInput) (User, error) {
	email := input.Email
	password := input.Password

	user, err := s.repository.FindByEmail(email)
	if err != nil {
		return user, err
	}

	if user.ID == 0 {
		return user, errors.New("User Not Found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *service) IsEmailAvailable(input CheckEmailInput) (bool, error) {
	email := input.Email

	user, err := s.repository.FindByEmail(email)
	if err != nil {
		return false, err
	}

	if user.ID == 0 {
		return true, nil
	}

	return false, nil
}

func (s *service) SaveAvatar(ID int, fileLocation string) (User, error) {
	user, err := s.repository.FIndByID(ID)
	if err != nil {
		return user, err
	}

	user.AvatarFileName = fileLocation

	updatedUser, err := s.repository.Update(user)
	if err != nil {
		return updatedUser, err
	}

	return updatedUser, nil
}

func (s *service) GetUserByID(ID int) (User, error) {
	user, err := s.repository.FIndByID(ID)
	if err != nil {
		return user, err
	}

	if user.ID == 0 {
		return user, errors.New("User Not Found")
	}
	return user, nil
}

//func (s *service) UploadToCloud(file *multipart.FileHeader, userId int) error {
//	bucket := "donation_alert"
//	object := file.Filename
//	pathName := fmt.Sprintf("%d-%s", userId, file.Filename)
//	ctx := context.Background()
//	client, err := storage.NewClient(ctx)
//	if err != nil {
//		return err
//	}
//	defer client.Close()
//
//	newFile, err := file.Open()
//	if err != nil {
//		return err
//	}
//	defer newFile.Close()
//	var b []byte
//
//	newFile.Read(b)
//	buf := bytes.NewBuffer(b)
//
//	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
//	defer cancel()
//
//	// Upload an object with storage2.Writer.
//	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
//	wc.ChunkSize = 0 // note retries are not supported for chunk size 0.
//
//	if _, err = io.Copy(wc, buf); err != nil {
//		return err
//	}
//	// Data can continue to be added to the file until the writer is closed.
//	if err := wc.Close(); err != nil {
//		return err
//	}
//	acl := client.Bucket(bucket).Object(object).ACL()
//	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
//		return err
//	}
//	return nil
//}
