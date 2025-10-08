import m from 'mithril'
import Modal from './Modal'
import { records } from '../services/api'

const RecordForm = {
  oninit(vnode) {
    this.initializeState()
    this.reset(vnode)
  },

  onbeforeupdate(vnode) {
    this.serverOptions = vnode.attrs.servers || []
    return true
  },

  onupdate(vnode) {
    if (this.prevContext !== vnode.attrs.context) {
      this.reset(vnode)
      this.prevContext = vnode.attrs.context
    }
  },

  initializeState() {
    this.fullDomain = ''
    this.recordType = 'CNAME'
    this.targetValue = ''
    this.ttl = 600
    this.notes = ''
    this.loading = false
    this.error = ''
    this.isEditMode = false
    this.selectedTargetServerId = null
    this.useCustomTarget = false
    this.hasCustomTarget = false
    this.serverOptions = []
    this.prevContext = null
    this.previousSelectedTargetServerId = null
  },

  reset(vnode) {
    const context = vnode?.attrs?.context
    this.serverOptions = vnode?.attrs?.servers || []
    this.loading = false
    this.error = ''
    this.hasCustomTarget = false
    this.previousSelectedTargetServerId = null

    if (context?.type === 'edit' && context.record) {
      const record = context.record
      this.isEditMode = true
      this.fullDomain = record.full_domain || ''
      this.recordType = record.record_type || 'CNAME'
      this.targetValue = record.target_value || ''
      this.ttl = record.ttl || 600
      this.notes = record.notes || ''
      const associatedId = context.associatedServerId
      const parsedId = associatedId === undefined || associatedId === null
        ? null
        : Number(associatedId)
      this.selectedTargetServerId = Number.isNaN(parsedId) ? null : parsedId
      if (this.selectedTargetServerId !== null) {
        this.useCustomTarget = false
        const server = this.getTargetServer()
        if (server) {
          const defaultTarget = this.recordType === 'CNAME'
            ? server.full_domain
            : server.target_value
          const normalizedRecordTarget = this.recordType === 'CNAME' && this.targetValue?.endsWith('.')
            ? this.targetValue.slice(0, -1)
            : this.targetValue
          const normalizedDefaultTarget = this.recordType === 'CNAME' && defaultTarget?.endsWith('.')
            ? defaultTarget.slice(0, -1)
            : defaultTarget
          this.hasCustomTarget = normalizedRecordTarget !== normalizedDefaultTarget
          if (!this.hasCustomTarget) {
            this.targetValue = defaultTarget
          }
        } else {
          this.useCustomTarget = true
          this.selectedTargetServerId = null
          this.hasCustomTarget = true
        }
      } else {
        this.useCustomTarget = true
        this.hasCustomTarget = true
      }
    } else {
      this.isEditMode = false
      this.fullDomain = ''
      this.recordType = 'CNAME'
      this.targetValue = ''
      this.ttl = 600
      this.notes = ''
      const initialId = context?.serverId ?? this.serverOptions[0]?.id ?? null
      this.selectedTargetServerId = initialId !== null ? Number(initialId) : null
      this.useCustomTarget = this.selectedTargetServerId === null
      this.applyDefaultTarget(true)
    }

    this.prevContext = context
  },

  getTargetServer() {
    if (this.selectedTargetServerId === null || this.selectedTargetServerId === undefined) return null
    return (this.serverOptions || []).find(server => server.id === this.selectedTargetServerId)
  },

  applyDefaultTarget(force = false) {
    if (this.useCustomTarget) {
      return
    }
    const server = this.getTargetServer()
    if (!server) return
    if (!this.hasCustomTarget || force) {
      this.targetValue = this.recordType === 'CNAME' ? server.full_domain : server.target_value
    }
  },

  async handleSubmit(vnode) {
    const context = vnode.attrs.context || {}
    this.error = ''
    this.loading = true

    try {
      if (this.isEditMode && context.record) {
        const record = context.record
        const payload = {
          provider_id: record.provider_id,
          zone_id: record.zone_id,
          zone_name: record.zone_name,
          full_domain: this.fullDomain,
          record_type: this.recordType,
          target_value: this.targetValue,
          ttl: this.ttl,
          notes: this.notes,
        }

        if (!this.useCustomTarget) {
          const selectedServer = this.getTargetServer()
          if (!selectedServer) {
            this.error = '请选择可用的指向目标'
            this.loading = false
            m.redraw()
            return
          }

          payload.target_value = this.recordType === 'CNAME'
            ? selectedServer.full_domain
            : selectedServer.target_value
        }

        await records.update(record.id, payload)
      } else {
        const payload = {
          full_domain: this.fullDomain,
          record_type: this.recordType,
          target_value: this.targetValue,
          ttl: this.ttl,
          notes: this.notes,
        }

        if (this.useCustomTarget) {
          if (!payload.target_value) {
            this.error = '请输入指向目标'
            this.loading = false
            m.redraw()
            return
          }
        } else {
          const selectedServer = this.getTargetServer()
          if (!selectedServer) {
            this.error = '请选择可用的指向目标'
            this.loading = false
            m.redraw()
            return
          }

          payload.target_value = this.recordType === 'CNAME'
            ? selectedServer.full_domain
            : selectedServer.target_value

          payload.provider_id = selectedServer.provider_id
          payload.zone_id = selectedServer.zone_id
          payload.zone_name = selectedServer.zone_name
        }

        await records.create(payload)
      }

      this.reset(vnode)
      vnode.attrs.onComplete && vnode.attrs.onComplete()
    } catch (error) {
      this.error = error.response?.error || (this.isEditMode ? '更新失败' : '创建失败')
    } finally {
      this.loading = false
      m.redraw()
    }
  },

  handleTargetServerChange(value) {
    const parsed = value === '' ? null : Number(value)
    if (parsed === null || Number.isNaN(parsed)) {
      this.previousSelectedTargetServerId = this.selectedTargetServerId
      this.selectedTargetServerId = null
      this.useCustomTarget = true
      return
    }

    this.previousSelectedTargetServerId = parsed
    this.selectedTargetServerId = parsed
    this.useCustomTarget = false
    this.hasCustomTarget = false
    this.applyDefaultTarget(true)
  },

  toggleCustomTarget() {
    const nextValue = !this.useCustomTarget
    this.useCustomTarget = nextValue

    if (nextValue) {
      this.previousSelectedTargetServerId = this.selectedTargetServerId
      this.selectedTargetServerId = null
      this.hasCustomTarget = true
    } else {
      const previousId = this.previousSelectedTargetServerId
      const hasPrevious = previousId !== null && this.serverOptions.some(server => server.id === previousId)
      const fallbackId = hasPrevious ? previousId : this.serverOptions[0]?.id || null
      this.selectedTargetServerId = fallbackId !== null ? Number(fallbackId) : null
      this.hasCustomTarget = false
      this.applyDefaultTarget(true)
    }
  },

  view(vnode) {
    const context = vnode.attrs.context || {}
    const onClose = vnode.attrs.onClose
    const targetServer = this.getTargetServer()

    const title = this.isEditMode && context.record
      ? `编辑记录: ${context.record.full_domain}`
      : '添加域名记录'

    const submitText = this.isEditMode
      ? (this.loading ? '保存中...' : '保存')
      : (this.loading ? '创建中...' : '添加')

    const hasServers = (this.serverOptions || []).length > 0

    const targetControl = () => {
      if (this.useCustomTarget || !hasServers) {
        return m('input', {
          type: 'text',
          value: this.targetValue,
          oninput: (e) => {
            this.targetValue = e.target.value
            this.hasCustomTarget = true
          },
          placeholder: this.isEditMode
            ? ''
            : targetServer
              ? (this.recordType === 'CNAME' ? targetServer.full_domain : targetServer.target_value)
              : '',
        })
      }

      return m('select', {
        value: this.selectedTargetServerId || '',
        onchange: (e) => this.handleTargetServerChange(e.target.value),
        disabled: this.loading,
      }, (this.serverOptions || []).map(server =>
        m('option', { value: server.id, key: `target-${server.id}` },
          `${server.server_name || server.full_domain} → ${this.recordType === 'CNAME' ? server.full_domain : server.target_value}`)
      ))
    }

    return m(Modal, {
      title,
      onClose: () => {
        this.reset(vnode)
        onClose && onClose()
      },
      footer: [
        m('button.btn.btn-secondary', {
          onclick: () => {
            this.reset(vnode)
            onClose && onClose()
          }
        }, '取消'),
        m('button.btn.btn-primary', {
          onclick: () => this.handleSubmit(vnode),
          disabled: this.loading || !this.fullDomain || (this.useCustomTarget && !this.targetValue)
        }, submitText),
      ]
    }, [
      m('.form-group', [
        m('label', '域名'),
        m('input', {
          type: 'text',
          value: this.fullDomain,
          oninput: (e) => { this.fullDomain = e.target.value },
          placeholder: 'app.example.com',
          autofocus: !this.isEditMode,
          disabled: this.isEditMode,
        })
      ]),

      m('.form-group', [
        m('label', '记录类型'),
        m('select', {
          value: this.recordType,
          onchange: (e) => {
            this.recordType = e.target.value
            if (!this.isEditMode) {
              this.applyDefaultTarget(true)
            }
          },
          disabled: this.isEditMode,
        }, [
          m('option', { value: 'CNAME' }, 'CNAME'),
          m('option', { value: 'A' }, 'A 记录'),
        ])
      ]),

      m('.form-group', [
        m('label', '指向'),
        targetControl()
      ]),

      hasServers && m('.form-group', {
        style: 'margin-top: -10px;'
      }, [
        m('label', {
          style: 'display: block; font-weight: 400; color: var(--text-gray); font-size: 13px;'
        }, [
          '如需自定义指向，可 ',
          m('a', {
            href: '#',
            onclick: (e) => {
              e.preventDefault()
              this.toggleCustomTarget()
            }
          }, this.useCustomTarget ? '恢复为服务器指向' : '切换为手动输入')
        ])
      ]),

      m('.form-group', [
        m('label', '备注（可选）'),
        m('input', {
          type: 'text',
          value: this.notes,
          oninput: (e) => { this.notes = e.target.value },
          placeholder: '例如：主博客'
        })
      ]),

      this.error && m('.error-message', this.error),
    ])
  }
}

export default RecordForm
