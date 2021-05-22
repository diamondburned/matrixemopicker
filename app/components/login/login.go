package login

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/chanbakjsd/gotrix"
	"github.com/chanbakjsd/gotrix/api"
	"github.com/chanbakjsd/gotrix/matrix"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed login.xml
var loginXML string

type Login struct {
	Done func(*gotrix.Client)

	builder       *gtk.Builder
	welcomePage   welcomePage
	passwordLogin passwordLogin

	client *gotrix.Client
}

type welcomePage struct {
	toplevel *gtk.Window

	homeserver      *gtk.Entry
	homeserverLabel *gtk.Label

	password *gtk.Button
	email    *gtk.Button
	token    *gtk.Button

	revealFunc func(bool)
}

type passwordLogin struct {
	dialog   *gtk.Dialog
	errLabel *gtk.Label

	email    *gtk.Entry
	password *gtk.Entry

	login *gtk.Button
	bound bool
}

func NewLogin(done func(*gotrix.Client)) *Login {
	b, err := gtk.BuilderNewFromString(loginXML)
	if err != nil {
		log.Fatalln("failed to build login.xml:", err)
	}

	login := Login{
		Done:    done,
		builder: b,
		welcomePage: welcomePage{
			toplevel:        builderMustGet(b, "toplevel").(*gtk.Window),
			homeserver:      builderMustGet(b, "homeserver-entry").(*gtk.Entry),
			homeserverLabel: builderMustGet(b, "homeserver-label").(*gtk.Label),
			password:        builderMustGet(b, "password-button").(*gtk.Button),
			email:           builderMustGet(b, "email-button").(*gtk.Button),
			token:           builderMustGet(b, "token-button").(*gtk.Button),
		},
		passwordLogin: passwordLogin{
			dialog:   builderMustGet(b, "password-dialog").(*gtk.Dialog),
			errLabel: builderMustGet(b, "password-errlabel").(*gtk.Label),
			email:    builderMustGet(b, "password-email").(*gtk.Entry),
			password: builderMustGet(b, "password-password").(*gtk.Entry),
			login:    builderMustGet(b, "password-login").(*gtk.Button),
		},
	}

	login.welcomePage.revealFunc = func(reveal bool) {
		login.welcomePage.password.SetSensitive(reveal)
		login.welcomePage.email.SetSensitive(reveal)
		login.welcomePage.token.SetSensitive(reveal)
	}
	login.welcomePage.revealFunc(false)

	login.welcomePage.homeserver.Connect("changed", func(e *gtk.Entry) {
		text, _ := e.GetText()
		login.welcomePage.revealFunc(text != "")
	})

	login.bindWelcomeButton(login.welcomePage.password, matrix.LoginPassword)
	login.bindWelcomeButton(login.welcomePage.email, matrix.LoginEmail)
	login.bindWelcomeButton(login.welcomePage.token, matrix.LoginToken)

	if hasSavedSession() {
		login.welcomePage.toplevel.SetSensitive(false)

		go func() {
			client, err := restoreSession()

			glib.IdleAdd(func() {
				if err != nil {
					login.welcomePage.toplevel.SetSensitive(true)
					return
				}

				login.done(client)
			})
		}()
	}

	return &login
}

func (l *Login) Show() {
	l.welcomePage.toplevel.SetSensitive(true)
	l.welcomePage.toplevel.ShowAll()
}

// Close closes the login dialog.
func (l *Login) Close() {
	l.welcomePage.toplevel.Close()
	l.passwordLogin.dialog.Close()
}

func (l *Login) done(c *gotrix.Client) {
	if err := saveSession(c); err != nil {
		log.Println("non-fatal error: cannot save session into keyring:", err)
	}

	l.Done(c)
	l.Close()
}

func (l *Login) bindWelcomeButton(b *gtk.Button, method matrix.LoginMethod) {
	b.Connect("clicked", func() {
		l.welcomePage.revealFunc(false)
		l.welcomePage.homeserver.SetSensitive(false)

		homeserver, _ := l.welcomePage.homeserver.GetText()

		go func() {
			c, err := gotrix.Discover(homeserver)

			glib.IdleAdd(func() {
				l.welcomePage.revealFunc(true)
				l.welcomePage.homeserver.SetSensitive(true)

				if err == nil {
					l.client = c
					l.moveOn(method)
					return
				}

				l.welcomePage.homeserverLabel.SetMarkup(fmt.Sprintf(
					`<span color="red" weight="bold">Error:</span> %v`, err,
				))
			})
		}()
	})
}

func (l *Login) moveOn(method matrix.LoginMethod) {
	switch method {
	case matrix.LoginPassword:
		l.spawnPasswordLogin()
	case matrix.LoginEmail:
		panic("unimplemented")
	case matrix.LoginToken:
		panic("unimplemented")
	}
}

func (l *Login) spawnPasswordLogin() {
	pl := &l.passwordLogin

	if !pl.bound {
		pl.bound = true
		pl.login.Connect("clicked", func() {
			pl.dialog.SetSensitive(false)

			client := l.client
			email, _ := pl.email.GetText()
			password, _ := pl.password.GetText()

			go func() {
				err := client.Login(api.LoginArg{
					Type:     matrix.LoginPassword,
					Password: password,
					Identifier: matrix.Identifier{
						Type: matrix.IdentifierUser,
						User: email,
					},
				})

				glib.IdleAdd(func() {
					pl.dialog.SetSensitive(true)

					if err == nil {
						pl.errLabel.SetLabel("")
						l.done(client)
						return
					}

					pl.errLabel.SetMarkup(fmt.Sprintf(
						`<span color="red" weight="bold">Error:</span> %v`, err,
					))
				})
			}()
		})
	}

	pl.dialog.ShowAll()
}

func builderMustGet(b *gtk.Builder, id string) glib.IObject {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatalln("cannot find id", id, "in login.xml")
	}

	return obj
}
