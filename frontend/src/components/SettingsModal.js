import m from 'mithril'
import { auth } from '../services/api'

const SettingsModal = {
  // State for username change
  newUsername: '',
  usernamePassword: '',
  usernameError: '',
  usernameSuccess: '',
  usernameLoading: false,

  // State for password change
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
  passwordError: '',
  passwordSuccess: '',
  passwordLoading: false,

  oninit(vnode) {
    this.currentUser = vnode.attrs.user
    this.onClose = vnode.attrs.onClose
    this.onUserUpdate = vnode.attrs.onUserUpdate
  },

  resetUsernameForm() {
    this.newUsername = ''
    this.usernamePassword = ''
    this.usernameError = ''
    this.usernameSuccess = ''
  },

  resetPasswordForm() {
    this.oldPassword = ''
    this.newPassword = ''
    this.confirmPassword = ''
    this.passwordError = ''
    this.passwordSuccess = ''
  },

  async handleChangeUsername() {
    this.usernameError = ''
    this.usernameSuccess = ''

    // Validation
    if (!this.newUsername) {
      this.usernameError = '请输入新用户名'
      return
    }

    if (this.newUsername.length < 3) {
      this.usernameError = '用户名至少需要 3 个字符'
      return
    }

    if (!this.usernamePassword) {
      this.usernameError = '请输入当前密码以确认身份'
      return
    }

    if (this.newUsername === this.currentUser.username) {
      this.usernameError = '新用户名与当前用户名相同'
      return
    }

    this.usernameLoading = true
    m.redraw()

    try {
      const response = await auth.changeUsername({
        new_username: this.newUsername,
        password: this.usernamePassword,
      })

      this.usernameSuccess = '用户名修改成功！'
      this.usernameError = ''

      // Update user in parent component
      if (this.onUserUpdate && response.user) {
        this.onUserUpdate(response.user)
      }

      // Reset form
      setTimeout(() => {
        this.resetUsernameForm()
        m.redraw()
      }, 2000)

    } catch (error) {
      console.error('Change username error:', error)
      this.usernameError = error.response?.error || error.message || '修改用户名失败'
      this.usernameSuccess = ''
    } finally {
      this.usernameLoading = false
      m.redraw()
    }
  },

  async handleChangePassword() {
    this.passwordError = ''
    this.passwordSuccess = ''

    // Validation
    if (!this.oldPassword) {
      this.passwordError = '请输入当前密码'
      return
    }

    if (!this.newPassword) {
      this.passwordError = '请输入新密码'
      return
    }

    if (this.newPassword.length < 6) {
      this.passwordError = '新密码至少需要 6 个字符'
      return
    }

    if (this.newPassword !== this.confirmPassword) {
      this.passwordError = '两次输入的新密码不一致'
      return
    }

    if (this.oldPassword === this.newPassword) {
      this.passwordError = '新密码不能与当前密码相同'
      return
    }

    this.passwordLoading = true
    m.redraw()

    try {
      await auth.changePassword({
        old_password: this.oldPassword,
        new_password: this.newPassword,
      })

      this.passwordSuccess = '密码修改成功！'
      this.passwordError = ''

      // Reset form
      setTimeout(() => {
        this.resetPasswordForm()
        m.redraw()
      }, 2000)

    } catch (error) {
      console.error('Change password error:', error)
      this.passwordError = error.response?.error || error.message || '修改密码失败'
      this.passwordSuccess = ''
    } finally {
      this.passwordLoading = false
      m.redraw()
    }
  },

  view() {
    return m('.modal-overlay', {
      onclick: (e) => {
        if (e.target.classList.contains('modal-overlay')) {
          this.onClose()
        }
      }
    }, [
      m('.modal', [
        m('.modal-header', [
          m('h2', '账户设置'),
          m('button.close-btn', {
            onclick: () => this.onClose()
          }, '×'),
        ]),

        m('.modal-body', [
          // Current user info
          m('.settings-section', [
            m('h3', '当前账户信息'),
            m('.info-row', [
              m('span.label', '用户名:'),
              m('span.value', this.currentUser?.username || ''),
            ]),
          ]),

          // Change Username Section
          m('.settings-section', [
            m('h3', '修改用户名'),

            m('.form-group', [
              m('label', '新用户名'),
              m('input.form-control', {
                type: 'text',
                placeholder: '请输入新用户名',
                value: this.newUsername,
                oninput: (e) => { this.newUsername = e.target.value },
                disabled: this.usernameLoading,
              }),
            ]),

            m('.form-group', [
              m('label', '当前密码（用于验证）'),
              m('input.form-control', {
                type: 'password',
                placeholder: '请输入当前密码',
                value: this.usernamePassword,
                oninput: (e) => { this.usernamePassword = e.target.value },
                disabled: this.usernameLoading,
              }),
            ]),

            this.usernameError && m('.alert.alert-error', this.usernameError),
            this.usernameSuccess && m('.alert.alert-success', this.usernameSuccess),

            m('.form-actions', [
              m('button.btn.btn-primary', {
                onclick: () => this.handleChangeUsername(),
                disabled: this.usernameLoading,
              }, this.usernameLoading ? '修改中...' : '修改用户名'),
            ]),
          ]),

          // Change Password Section
          m('.settings-section', [
            m('h3', '修改密码'),

            m('.form-group', [
              m('label', '当前密码'),
              m('input.form-control', {
                type: 'password',
                placeholder: '请输入当前密码',
                value: this.oldPassword,
                oninput: (e) => { this.oldPassword = e.target.value },
                disabled: this.passwordLoading,
              }),
            ]),

            m('.form-group', [
              m('label', '新密码'),
              m('input.form-control', {
                type: 'password',
                placeholder: '请输入新密码（至少6位）',
                value: this.newPassword,
                oninput: (e) => { this.newPassword = e.target.value },
                disabled: this.passwordLoading,
              }),
            ]),

            m('.form-group', [
              m('label', '确认新密码'),
              m('input.form-control', {
                type: 'password',
                placeholder: '请再次输入新密码',
                value: this.confirmPassword,
                oninput: (e) => { this.confirmPassword = e.target.value },
                disabled: this.passwordLoading,
              }),
            ]),

            this.passwordError && m('.alert.alert-error', this.passwordError),
            this.passwordSuccess && m('.alert.alert-success', this.passwordSuccess),

            m('.form-actions', [
              m('button.btn.btn-primary', {
                onclick: () => this.handleChangePassword(),
                disabled: this.passwordLoading,
              }, this.passwordLoading ? '修改中...' : '修改密码'),
            ]),
          ]),
        ]),
      ]),
    ])
  },
}

export default SettingsModal
