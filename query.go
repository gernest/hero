package hero

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type query struct {
	*gorm.DB
}

func (q *query) ClientByCode(code string) (*Client, error) {
	c := &Client{}
	if code == "" {
		return nil, errors.New("invalid code")
	}
	d := q.Where(&Client{UUID: code}).First(c)
	if d.Error != nil {
		return nil, d.Error
	}
	return c, nil

}

func (q *query) GrantByRefreshToken(code string) (*Grant, error) {
	tok := &Token{}
	d := q.Where(&Token{Code: code}).First(tok)
	if d.Error != nil {
		return nil, d.Error
	}
	g := &Grant{}
	d = q.Where(&Grant{RefreshTokenID: tok.ID}).First(g)
	if d.Error != nil {
		return nil, d.Error
	}
	return g, nil
}

func (q *query) GrantByCode(code string) (*Grant, error) {
	g := &Grant{}
	d := q.Where(&Grant{Code: code}).Preload("AccessToken").Preload("AuthorizaToken").
		Preload("RefreshToken").First(g)
	if d.Error != nil {
		return nil, d.Error
	}
	return g, nil
}

func (q *query) GrantByCLient(c *Client, code string) (*Grant, error) {
	g := &Grant{}
	d := q.Where(&Grant{Code: code, ClientID: c.ID}).Preload("AccessToken").Preload("AuthorizeToken").
		Preload("RefreshToken").First(g)
	if d.Error != nil {
		return nil, d.Error
	}
	return g, nil
}

func (q *query) GrantByBearer(bearerCode string) (*Grant, error) {
	tok, err := q.TokenByCode(bearerCode)
	if err != nil {
		return nil, err
	}
	g := &Grant{}
	d := q.Where(&Grant{AccessTokenID: tok.ID}).
		Preload("AccessToken").Preload("RefreshToken").First(g)
	if d.Error != nil {
		return nil, d.Error
	}
	return g, nil
}

func (q *query) SaveModel(model interface{}) error {
	return q.Save(model).Error
}

func (q *query) UpdateModel(model interface{}) error {
	return q.Update(model).Error
}

func (q *query) DeleteModel(model interface{}) error {
	return q.Delete(model).Error
}

func (q *query) GetSessionByKey(key string) (*Session, error) {
	ss := &Session{}
	d := q.Where(&Session{Key: key}).First(ss)
	if d.Error != nil {
		return nil, d.Error
	}
	return ss, nil
}

func (q *query) UpdateSession(sess *Session) error {
	ss, err := q.GetSessionByKey(sess.Key)
	if err != nil {
		return err
	}
	ss.Data = sess.Data
	return q.Save(ss).Error
}

func (q *query) DeleteSession(key string) error {
	ss, err := q.GetSessionByKey(key)
	if err != nil {
		return err
	}
	return q.Delete(ss).Error
}

func (q *query) SaveSession(ss *Session) error {
	return q.Save(ss).Error
}

func (q *query) UserByID(id int64) (*User, error) {
	usr := &User{}
	d := q.Where(&User{ID: id}).First(usr)
	if d.Error != nil {
		return nil, d.Error
	}
	return usr, nil
}

func (q *query) UserByUserName(username string) (*User, error) {
	usr := &User{}
	d := q.Where(&User{UserName: username}).First(usr)
	if d.Error != nil {
		return nil, d.Error
	}
	return usr, nil
}

func (q *query) UserByEmail(email string) (*User, error) {
	usr := &User{}
	d := q.Where(&User{Email: email}).First(usr)
	if d.Error != nil {
		return nil, d.Error
	}
	return usr, nil
}

func (q *query) CreateUser(usr *User) error {
	return q.Save(usr).Error
}

func (q *query) TokenByCode(code string) (*Token, error) {
	tok := &Token{}
	d := q.Where(&Token{Code: code}).First(tok)
	if d.Error != nil {
		return nil, d.Error
	}
	return tok, nil
}
