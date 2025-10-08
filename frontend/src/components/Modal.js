import m from 'mithril'

const Modal = {
  view(vnode) {
    const { title, onClose, footer, className } = vnode.attrs
    const modalClass = className ? `modal ${className}` : 'modal'

    return m('.modal-overlay', {
      onclick: (e) => {
        if (e.target.classList.contains('modal-overlay')) {
          onClose && onClose()
        }
      }
    }, [
      m('div', { class: modalClass }, [
        m('.modal-header', [
          m('h3', title),
          m('button.modal-close', {
            onclick: onClose
          }, '×')
        ]),
        m('.modal-body', vnode.children),
        footer && m('.modal-footer', footer),
      ])
    ])
  }
}

export default Modal
