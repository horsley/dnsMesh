import m from 'mithril'
import { auth, records } from '../services/api'
import ProviderWizard from '../components/ProviderWizard'
import RecordForm from '../components/RecordForm'
import SettingsModal from '../components/SettingsModal'
import AuditLogModal from '../components/AuditLogModal'

const Dashboard = {
  user: null,
  servers: [],
  unassignedRecords: [],
  loading: true,
  showProviderWizard: false,
  showSettings: false,
  showRecordForm: false,
  showAuditLogs: false,
  recordFormContext: null,
  reanalyzing: false,
  activeMenuId: null,

  oninit() {
    this.loadData()
  },

  async loadData() {
    this.loading = true

    try {
      // Load user info
      const userStr = localStorage.getItem('user')
      if (userStr) {
        this.user = JSON.parse(userStr)
      }

      // Load records (new server-first structure)
      const response = await records.list()
      this.servers = response.servers || []
      this.unassignedRecords = response.unassigned_records || []
    } catch (error) {
      console.error('Failed to load data:', error)
      if (error.code === 401) {
        m.route.set('/login')
      }
    } finally {
      this.loading = false
      m.redraw()
    }
  },

  async handleHideRecord(recordId) {
    if (!confirm('确定要隐藏此记录吗？隐藏后将脱离系统管理。')) return

    try {
      await records.hide(recordId)
      await this.loadData()
    } catch (error) {
      alert('隐藏失败: ' + (error.response?.error || error.message))
    }
  },

  async handleDeleteRecord(recordId) {
    if (!confirm('确定要删除此记录吗？此操作将从 DNS Provider 中删除记录，无法恢复！')) return

    try {
      await records.delete(recordId)
      await this.loadData()
    } catch (error) {
      alert('删除失败: ' + (error.response?.error || error.message))
    }
  },

  async handleLogout() {
    try {
      await auth.logout()
      localStorage.removeItem('user')
      m.route.set('/login')
    } catch (error) {
      console.error('Logout failed:', error)
    }
  },

  async handleReanalyze() {
    if (!confirm('确定要重新分析所有记录吗？这将从所有 Provider 重新同步记录并更新服务器分组。')) return

    this.reanalyzing = true
    m.redraw()

    try {
      await records.reanalyze()
      await this.loadData()
      alert('重新分析完成！')
    } catch (error) {
      alert('重新分析失败: ' + (error.response?.error || error.message))
    } finally {
      this.reanalyzing = false
      m.redraw()
    }
  },

  openRecordForm(context) {
    this.closeMenu()
    this.recordFormContext = context
    this.showRecordForm = true
  },

  closeRecordForm() {
    this.showRecordForm = false
    this.recordFormContext = null
  },

  toggleMenu(id) {
    this.activeMenuId = this.activeMenuId === id ? null : id
  },

  closeMenu() {
    this.activeMenuId = null
  },

  buildDomainUrl(domain) {
    if (!domain) return '#'
    const trimmed = domain.trim()
    if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) {
      return trimmed
    }
    return `https://${trimmed}`
  },

  handleUserUpdate(updatedUser) {
    // Update user in component state
    this.user = updatedUser
    // Update user in localStorage
    localStorage.setItem('user', JSON.stringify(updatedUser))
    m.redraw()
  },

  view() {
    return [
      // Header
      m('.header', {
        onclick: () => this.closeMenu()
      }, [
        m('.header-content', [
          m('h1', 'DNSMesh'),
          m('.header-actions', [
            m('.header-user', `欢迎, ${this.user?.username || 'User'}`),
            m('button.btn.btn-secondary.btn-small', {
              onclick: () => { this.showSettings = true },
              style: 'margin-right: 10px;'
            }, '⚙️ 设置'),
            m('button.btn.btn-secondary.btn-small', {
              onclick: () => this.handleLogout()
            }, '退出登录'),
          ]),
        ]),
      ]),

      // Main Content
      m('.container', [
        m('.main-content', {
          onclick: () => this.closeMenu()
        }, [
          m('.toolbar', [
            m('.toolbar-left', [
              m('button.btn.btn-primary.btn-small', {
                onclick: () => {
                  const primaryServers = (this.servers || []).map(group => group.server)
                  if (!primaryServers.length) {
                    alert('请先同步服务器记录后再添加域名。')
                    return
                  }
                  this.openRecordForm({ type: 'create' })
                }
              }, '+ 添加域名'),
              m('button.btn.btn-secondary.btn-small', {
                onclick: () => this.handleReanalyze(),
                disabled: this.reanalyzing,
                style: 'white-space: nowrap;'
              }, this.reanalyzing ? '分析中...' : '🔄 重新分析'),
              m('button.btn.btn-secondary.btn-small', {
                onclick: () => { this.showProviderWizard = true }
              }, '+ 添加 Provider'),
            ]),
            m('.toolbar-right', [
              m('button.btn.btn-secondary.btn-small', {
                onclick: () => { this.showAuditLogs = true }
              }, '📜 审计日志'),
            ]),
          ]),

          this.loading ?
            m('.loading', '加载中...') :
            this.servers.length === 0 && this.unassignedRecords.length === 0 ?
              m('.empty-state', [
                m('.empty-state-icon', '📋'),
                m('p', '暂无数据'),
                m('p', '请先添加 DNS Provider'),
              ]) :
              this.renderRecords(),
        ]),
      ]),

      // Modals
      this.showProviderWizard && m(ProviderWizard, {
        onClose: () => { this.showProviderWizard = false },
        onComplete: () => {
          this.showProviderWizard = false
          this.loadData()
        }
      }),

      this.showSettings && m(SettingsModal, {
        user: this.user,
        onClose: () => { this.showSettings = false },
        onUserUpdate: (user) => this.handleUserUpdate(user),
      }),

      this.showRecordForm && m(RecordForm, {
        context: this.recordFormContext,
        servers: (this.servers || []).map(group => group.server),
        onClose: () => {
          this.closeRecordForm()
        },
        onComplete: () => {
          this.closeRecordForm()
          this.loadData()
        },
      }),

      this.showAuditLogs && m(AuditLogModal, {
        onClose: () => { this.showAuditLogs = false }
      }),
    ]
  },

  renderRecords() {
    return [
      // Render all servers (no provider grouping)
      this.servers.map(serverGroup =>
        m('.server-card', { key: serverGroup.server.id }, [
          m('.server-header', [
            m('.server-info', [
              m('span.server-icon', '🖥️'),
              m('span.server-name', [
                serverGroup.server.server_name || serverGroup.server.full_domain,
                m('a.record-domain-link', {
                  href: this.buildDomainUrl(serverGroup.server.full_domain),
                  target: '_blank',
                  rel: 'noopener noreferrer',
                  title: '在新标签页打开'
                }, '↗'),
              ]),
              m('span.server-meta', [
                serverGroup.server.server_region && m('span', serverGroup.server.server_region + ' '),
                m('a', {
                  class: 'record-domain-inline',
                  href: this.buildDomainUrl(serverGroup.server.full_domain),
                  target: '_blank',
                  rel: 'noopener noreferrer'
                }, serverGroup.server.full_domain),
                m('span', ' → ' + serverGroup.server.target_value),
              ]),
            ]),
            m('.server-actions', [
              m('button.action-trigger', {
                onclick: (e) => {
                  e.stopPropagation()
                  this.toggleMenu(`server-${serverGroup.server.id}`)
                }
              }, '⋯'),
              this.activeMenuId === `server-${serverGroup.server.id}` &&
                m('.action-menu', {
                  onclick: (e) => e.stopPropagation(),
                }, [
                  m('button.action-item', {
                    onclick: () => {
                      this.closeMenu()
                      this.openRecordForm({ type: 'create', serverId: serverGroup.server.id })
                    }
                  }, '添加关联域名'),
                  m('button.action-item.danger', {
                    onclick: () => {
                      this.closeMenu()
                      this.handleHideRecord(serverGroup.server.id)
                    }
                  }, '隐藏服务器'),
                ]),
            ]),
          ]),

          // Related records
          serverGroup.related_records?.length > 0 && m('.record-list',
            serverGroup.related_records.map(record =>
              m('.record-item', { key: record.id }, [
                m('.record-info', [
                  m('.record-title', [
                    m('a', {
                      class: 'record-domain',
                      href: this.buildDomainUrl(record.full_domain),
                      target: '_blank',
                      rel: 'noopener noreferrer'
                    }, record.full_domain),
                    m('span.record-type', record.record_type),
                    m('span.record-target', '→ ' + record.target_value),
                  ]),
                  record.notes && m('span.record-notes', record.notes),
                ]),
                this.renderRecordActions(record, { associatedServerId: serverGroup.server.id }),
              ])
            )
          ),
        ])
      ),

      // Render unassigned records grouped by provider
      this.unassignedRecords.length > 0 && [
        m('h3', { style: 'margin: 30px 0 15px 0; padding: 0 20px;' }, '未分组记录'),
        this.unassignedRecords.map(group =>
          m('.provider-section', { key: group.provider_id }, [
            m('.provider-header', [
              m('.provider-title', [
                m('.provider-icon', group.provider_name === 'cloudflare' ? '☁️' : '🌐'),
                m('span', group.provider_name === 'cloudflare' ? 'Cloudflare' : '腾讯云 DNSPod'),
              ]),
            ]),
            m('.record-list',
              group.records.map(record =>
                m('.record-item', { key: record.id }, [
                  m('.record-info', [
                    m('.record-title', [
                      m('a', {
                        class: 'record-domain',
                        href: this.buildDomainUrl(record.full_domain),
                        target: '_blank',
                        rel: 'noopener noreferrer'
                      }, record.full_domain),
                      m('span.record-type', record.record_type),
                      m('span.record-target', '→ ' + record.target_value),
                    ]),
                    record.notes && m('span.record-notes', record.notes),
                  ]),
                  this.renderRecordActions(record),
                ])
              )
            ),
          ])
        ),
      ],
    ]
  },

  renderRecordActions(record, options = {}) {
    const menuId = `record-${record.id}`
    return m('.record-actions', [
      m('button.action-trigger', {
        onclick: (e) => {
          e.stopPropagation()
          this.toggleMenu(menuId)
        }
      }, '⋯'),
      this.activeMenuId === menuId &&
        m('.action-menu', {
          onclick: (e) => e.stopPropagation(),
        }, [
          m('button.action-item', {
            onclick: () => {
              this.closeMenu()
              this.openRecordForm({
                type: 'edit',
                record,
                associatedServerId: options.associatedServerId ?? null,
              })
            }
          }, '编辑'),
          m('button.action-item', {
            onclick: () => {
              this.closeMenu()
              this.handleHideRecord(record.id)
            }
          }, '隐藏'),
          !record.is_server && m('button.action-item.danger', {
            onclick: () => {
              this.closeMenu()
              this.handleDeleteRecord(record.id)
            }
          }, '删除'),
        ]),
    ])
  },
}

export default Dashboard
