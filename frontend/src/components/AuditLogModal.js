import m from 'mithril'
import Modal from './Modal'
import { auditLogs } from '../services/api'

const ACTION_OPTIONS = [
  { label: '全部动作', value: '' },
  { label: '创建', value: 'create' },
  { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' },
  { label: '同步', value: 'sync' },
]

const RESOURCE_OPTIONS = [
  { label: '全部资源', value: '' },
  { label: 'Provider', value: 'provider' },
  { label: '解析记录', value: 'record' },
]

const AuditLogModal = {
  logs: [],
  loading: false,
  error: '',
  hasMore: false,
  limit: 20,
  offset: 0,
  filters: {
    action: '',
    resourceType: '',
  },

  oninit() {
    this.loadLogs(true)
  },

  resetAndLoad() {
    this.offset = 0
    this.logs = []
    this.loadLogs(true)
  },

  async loadLogs(reset = false) {
    if (this.loading) return
    this.loading = true
    this.error = ''
    m.redraw()

    try {
      const params = {
        limit: this.limit,
        offset: this.offset,
      }

      if (this.filters.action) {
        params.action = this.filters.action
      }

      if (this.filters.resourceType) {
        params.resource_type = this.filters.resourceType
      }

      const response = await auditLogs.list(params)
      const fetchedLogs = response.logs || []

      if (reset) {
        this.logs = fetchedLogs
      } else {
        this.logs = this.logs.concat(fetchedLogs)
      }

      this.hasMore = (response.count || 0) === this.limit

      if (this.hasMore) {
        this.offset += this.limit
      }
    } catch (error) {
      console.error('Failed to load audit logs:', error)
      this.hasMore = false
      this.error = error.response?.error || error.message || '获取审计日志失败'
    } finally {
      this.loading = false
      m.redraw()
    }
  },

  parseDetails(details) {
    if (!details) return '-'
    try {
      const parsed = JSON.parse(details)
      return Object.keys(parsed).length
        ? JSON.stringify(parsed, null, 2)
        : '-'
    } catch (error) {
      return details
    }
  },

  formatTimestamp(timestamp) {
    if (!timestamp) return '-'
    const date = new Date(timestamp)
    if (Number.isNaN(date.getTime())) return timestamp
    return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`
  },

  view(vnode) {
    const { onClose } = vnode.attrs

    return m(Modal, {
      title: '审计日志',
      onClose,
      className: 'modal-large',
      footer: [
        m('button.btn.btn-secondary.btn-small', { onclick: onClose }, '关闭')
      ],
    }, [
      m('.audit-filters', [
        m('div.filter-group', [
          m('label', '动作'),
          m('select', {
            value: this.filters.action,
            onchange: (e) => {
              this.filters.action = e.target.value
              this.resetAndLoad()
            },
          }, ACTION_OPTIONS.map((option) =>
            m('option', { value: option.value }, option.label)
          )),
        ]),
        m('div.filter-group', [
          m('label', '资源类型'),
          m('select', {
            value: this.filters.resourceType,
            onchange: (e) => {
              this.filters.resourceType = e.target.value
              this.resetAndLoad()
            },
          }, RESOURCE_OPTIONS.map((option) =>
            m('option', { value: option.value }, option.label)
          )),
        ]),
      ]),

      this.error && m('.audit-error', this.error),

      this.loading && this.logs.length === 0 ?
        m('.audit-loading', '加载中...') :
        this.logs.length === 0 ?
          m('.audit-empty', '暂无审计记录') :
          m('div.audit-table-wrapper', [
            m('table.audit-table', [
              m('thead', [
                m('tr', [
                  m('th', '时间'),
                  m('th', '动作'),
                  m('th', '资源类型'),
                  m('th', '资源 ID'),
                  m('th', '来源 IP'),
                  m('th', '详情'),
                ]),
              ]),
              m('tbody', this.logs.map((log) =>
                m('tr', { key: log.id }, [
                  m('td', this.formatTimestamp(log.created_at)),
                  m('td', log.action || '-'),
                  m('td', log.resource_type || '-'),
                  m('td', log.resource_id || '-'),
                  m('td', log.ip_address || '-'),
                  m('td', [
                    m('pre.audit-details', this.parseDetails(log.details)),
                  ]),
                ])
              )),
            ]),

            this.hasMore && m('div.audit-load-more', [
              m('button.btn.btn-secondary.btn-small', {
                onclick: () => this.loadLogs(),
                disabled: this.loading,
              }, this.loading ? '加载中...' : '加载更多')
            ]),
          ]),
    ])
  },
}

export default AuditLogModal
