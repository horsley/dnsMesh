import m from 'mithril'
import Modal from './Modal'
import { providers } from '../services/api'

const ProviderWizard = {
  step: 1,
  providerType: '',
  apiKey: '',
  apiSecret: '',
  loading: false,
  error: '',
  syncResult: null,
  selectedRecords: {},

  oninit() {
    this.reset()
  },

  reset() {
    this.step = 1
    this.providerType = ''
    this.apiKey = ''
    this.apiSecret = ''
    this.loading = false
    this.error = ''
    this.syncResult = null
    this.selectedRecords = {}
  },

  async handleConnect() {
    this.error = ''
    this.loading = true

    try {
      // Create provider
      const response = await providers.create({
        name: this.providerType,
        api_key: this.apiKey,
        api_secret: this.apiSecret,
      })

      // Sync records
      const syncResponse = await providers.sync(response.provider.id)
      this.syncResult = syncResponse
      this.syncResult.providerId = response.provider.id

      // Pre-select server suggestions
      if (syncResponse.server_suggestions) {
        syncResponse.server_suggestions.forEach(suggestion => {
          this.selectedRecords[suggestion.domain || suggestion.ip] = true

          // Select related records
          if (suggestion.referenced_by) {
            suggestion.referenced_by.forEach(domain => {
              this.selectedRecords[domain] = true
            })
          }
          if (suggestion.same_ip_domains) {
            suggestion.same_ip_domains.forEach(domain => {
              this.selectedRecords[domain] = true
            })
          }
        })
      }

      this.step = 2
    } catch (error) {
      this.error = error.response?.error || '连接失败'
    } finally {
      this.loading = false
      m.redraw()
    }
  },

  async handleImport(onComplete) {
    this.loading = true
    this.error = ''

    try {
      const recordsToImport = []

      console.log('=== Import Debug ===')
      console.log('selectedRecords:', this.selectedRecords)
      console.log('server_suggestions:', this.syncResult.server_suggestions)
      console.log('records count:', this.syncResult.records?.length)

      // Import selected server suggestions
      if (this.syncResult.server_suggestions) {
        this.syncResult.server_suggestions.forEach(suggestion => {
          const key = suggestion.domain || suggestion.ip
          console.log('Checking suggestion:', key, 'selected:', this.selectedRecords[key])

          if (!this.selectedRecords[key]) return

          const record = this.syncResult.records.find(r =>
            r.full_domain === suggestion.domain || r.target_value === suggestion.ip
          )

          console.log('Found record for suggestion:', record)

          if (record) {
            recordsToImport.push({
              zone_id: record.zone_id,
              zone_name: record.zone_name,
              full_domain: record.full_domain,
              record_type: record.record_type,
              target_value: record.target_value,
              ttl: record.ttl,
              provider_record_id: record.provider_record_id,
              is_server: true,
              server_name: suggestion.suggested_name || '',
              server_region: suggestion.suggested_region || '',
            })
          }
        })
      }

      // Import other selected records
      this.syncResult.records.forEach(record => {
        console.log('Checking record:', record.full_domain, 'selected:', this.selectedRecords[record.full_domain])

        if (!this.selectedRecords[record.full_domain]) return

        // Skip if already added as server
        if (recordsToImport.find(r => r.full_domain === record.full_domain)) return

        recordsToImport.push({
          zone_id: record.zone_id,
          zone_name: record.zone_name,
          full_domain: record.full_domain,
          record_type: record.record_type,
          target_value: record.target_value,
          ttl: record.ttl,
          provider_record_id: record.provider_record_id,
          is_server: false,
        })
      })

      console.log('recordsToImport count:', recordsToImport.length)
      console.log('recordsToImport:', recordsToImport)

      // Import records
      await m.request({
        method: 'POST',
        url: '/api/records/import',
        body: {
          provider_id: this.syncResult.providerId,
          records: recordsToImport,
        },
        withCredentials: true,
      })

      // Success, close modal and reload
      onComplete && onComplete()
    } catch (error) {
      this.error = error.response?.error || '导入失败'
    } finally {
      this.loading = false
      m.redraw()
    }
  },

  view(vnode) {
    const { onClose } = vnode.attrs

    if (this.step === 1) {
      return m(Modal, {
        title: '添加 DNS 提供商',
        onClose: () => {
          this.reset()
          onClose && onClose()
        },
        footer: [
          m('button.btn.btn-secondary', {
            onclick: () => {
              this.reset()
              onClose && onClose()
            }
          }, '取消'),
          m('button.btn.btn-primary', {
            onclick: () => this.handleConnect(),
            disabled: !this.providerType || !this.apiKey ||
                      (this.providerType === 'tencentcloud' && !this.apiSecret) ||
                      this.loading
          }, this.loading ? '连接中...' : '连接并同步'),
        ]
      }, [
        m('.form-group', [
          m('label', '提供商类型'),
          m('select', {
            value: this.providerType,
            onchange: (e) => { this.providerType = e.target.value }
          }, [
            m('option', { value: '' }, '请选择'),
            m('option', { value: 'cloudflare' }, 'Cloudflare'),
            m('option', { value: 'tencentcloud' }, '腾讯云 DNSPod'),
          ])
        ]),

        this.providerType === 'cloudflare' && [
          m('.form-group', [
            m('label', 'API Token'),
            m('input', {
              type: 'text',
              value: this.apiKey,
              oninput: (e) => { this.apiKey = e.target.value },
              placeholder: '在 Cloudflare 仪表板创建 API Token'
            })
          ]),
          m('p.form-help', { style: 'margin: -8px 0 12px 0; font-size: 12px; color: #666;' },
            '需要权限: Zone.Zone:Read, Zone.DNS:Edit'
          ),
        ],

        this.providerType === 'tencentcloud' && [
          m('.form-group', [
            m('label', 'Secret ID'),
            m('input', {
              type: 'text',
              value: this.apiKey,
              oninput: (e) => { this.apiKey = e.target.value },
              placeholder: 'AKID...'
            })
          ]),
          m('.form-group', [
            m('label', 'Secret Key'),
            m('input', {
              type: 'password',
              value: this.apiSecret,
              oninput: (e) => { this.apiSecret = e.target.value },
              placeholder: 'Secret Key'
            })
          ]),
        ],

        this.error && m('.error-message', this.error),
      ])
    }

    // Step 2: Import wizard
    return m(Modal, {
      title: '选择要导入的记录',
      onClose: () => {
        this.reset()
        onClose && onClose()
      },
      footer: [
        m('button.btn.btn-secondary', {
          onclick: () => {
            this.reset()
            onClose && onClose()
          }
        }, '取消'),
        m('button.btn.btn-primary', {
          onclick: () => this.handleImport(vnode.attrs.onComplete),
          disabled: this.loading
        }, this.loading ? '导入中...' : '导入选中的记录'),
      ]
    }, [
      m('p', `已分析 ${this.syncResult?.records?.length || 0} 条记录`),

      this.syncResult?.server_suggestions?.length > 0 && [
        m('h4', '💡 建议的服务器'),
        this.syncResult.server_suggestions.map(suggestion => {
          const key = suggestion.domain || suggestion.ip
          return m('.server-suggestion', { key }, [
            m('label', [
              m('input[type=checkbox]', {
                checked: this.selectedRecords[key],
                onchange: (e) => {
                  this.selectedRecords[key] = e.target.checked

                  // Toggle related records
                  if (suggestion.referenced_by) {
                    suggestion.referenced_by.forEach(d => {
                      this.selectedRecords[d] = e.target.checked
                    })
                  }
                  if (suggestion.same_ip_domains) {
                    suggestion.same_ip_domains.forEach(d => {
                      this.selectedRecords[d] = e.target.checked
                    })
                  }
                }
              }),
              m('strong', suggestion.domain || suggestion.ip),
              m('span', ` (${suggestion.confidence} - ${suggestion.match_reason})`),
            ]),

            suggestion.suggested_name && m('.server-details', [
              m('input', {
                placeholder: '服务器名称',
                value: suggestion.suggested_name,
                oninput: (e) => { suggestion.suggested_name = e.target.value }
              }),
              m('input', {
                placeholder: '地区',
                value: suggestion.suggested_region || '',
                oninput: (e) => { suggestion.suggested_region = e.target.value }
              }),
            ]),

            // Show related records
            (suggestion.referenced_by?.length > 0 || suggestion.same_ip_domains?.length > 0) &&
            m('.related-records', [
              suggestion.referenced_by?.map(domain =>
                m('div', { key: domain }, `  ├─ ${domain}`)
              ),
              suggestion.same_ip_domains?.map(domain =>
                m('div', { key: domain }, `  ├─ ${domain} (同IP)`)
              ),
            ])
          ])
        })
      ],

      this.error && m('.error-message', this.error),
    ])
  }
}

export default ProviderWizard
