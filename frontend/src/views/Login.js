import m from 'mithril'
import { auth } from '../services/api'

const Login = {
  username: '',
  password: '',
  error: '',
  loading: false,

  handleLogin() {
    this.error = ''
    this.loading = true

    auth
      .login({
        username: this.username,
        password: this.password,
      })
      .then((response) => {
        localStorage.setItem('user', JSON.stringify(response.user))
        m.route.set('/dashboard')
      })
      .catch((error) => {
        this.error = error.response?.error || '登录失败，请检查用户名和密码'
      })
      .finally(() => {
        this.loading = false
        m.redraw()
      })
  },

  view() {
    return m('.login-page', [
      m('.login-box', [
        m('h2', 'DNSMesh'),
        m('form', {
          onsubmit: (e) => {
            e.preventDefault()
            this.handleLogin()
          }
        }, [
          m('.form-group', [
            m('label', { for: 'username' }, '用户名'),
            m('input#username', {
              type: 'text',
              value: this.username,
              oninput: (e) => { this.username = e.target.value },
              placeholder: '请输入用户名',
              disabled: this.loading,
            }),
          ]),
          m('.form-group', [
            m('label', { for: 'password' }, '密码'),
            m('input#password', {
              type: 'password',
              value: this.password,
              oninput: (e) => { this.password = e.target.value },
              placeholder: '请输入密码',
              disabled: this.loading,
            }),
          ]),
          m('button.btn.btn-primary.btn-block[type=submit]', {
            disabled: this.loading
          }, this.loading ? '登录中...' : '登录'),
          this.error && m('.error-message', this.error),
        ]),
      ]),
    ])
  },
}

export default Login
