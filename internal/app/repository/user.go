package repository

import (
	"fmt"
	"lab4/internal/app/ds"
)

func (r *Repository) RegisterUser(input ds.ChangeUserDTO) error {
	var existingUser ds.User
	err := r.db.Where("login = ?", input.Login).First(&existingUser).Error
	if err == nil {
		return fmt.Errorf("пользователь с таким логином уже существует")
	}

	user := ds.User{
		Login:    input.Login,
		Password: input.Password,
	}
	return r.db.Create(&user).Error
}

func (r *Repository) LoginUser(login, password string) (*ds.UserDTO, error) {
	var user ds.User
	err := r.db.Where("login = ? AND password = ?", login, password).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("неверный логин или пароль")
	}

	role := ds.RoleCreator
	if user.IsModerator {
		role = ds.RoleModerator
	}

	return &ds.UserDTO{
		UserID: user.UserID,
		Login:  user.Login,
		Role:   role,
	}, nil
}

func (r *Repository) GetUserByID(userID int) (*ds.UserDTO, error) {
	var user ds.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	role := ds.RoleCreator
	if user.IsModerator {
		role = ds.RoleModerator
	}
	return &ds.UserDTO{
		UserID: user.UserID,
		Login:  user.Login,
		Role:   role,
	}, nil
}

func (r *Repository) UpdateUser(userID int, userUpdates ds.ChangeUserDTO) error {
	var user ds.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return err
	}

	if userUpdates.Login != "" {
		var count int64
		r.db.Model(&ds.User{}).Where("login = ? AND user_id != ?", userUpdates.Login, userID).Count(&count)
		if count > 0 {
			return fmt.Errorf("логин уже занят")
		}
		user.Login = userUpdates.Login
	}

	if userUpdates.Password != "" {
		user.Password = userUpdates.Password
	}

	return r.db.Save(&user).Error
}
