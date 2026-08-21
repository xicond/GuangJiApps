/**
 * Utility to scroll page or modal dialog to the top-most invalid form field / error element + 15px
 * when form validation errors or marked server-side field errors occur.
 */
export function scrollToFormError(containerElement?: HTMLElement | null, offset = 15) {
  setTimeout(() => {
    // 1. Find the top-most error element in DOM (e.g. .el-form-item.is-error, .is-error, .validation-alert)
    const scope = containerElement || document
    const firstErrorEl = scope.querySelector(
      '.el-form-item.is-error, .is-error, .validation-alert, .field-errors-container'
    ) as HTMLElement | null

    if (firstErrorEl) {
      const dialogBody = firstErrorEl.closest('.el-dialog__body') as HTMLElement | null
      if (dialogBody) {
        // Scroll within el-dialog container to top-most error field - 15px offset
        const dialogRect = dialogBody.getBoundingClientRect()
        const errorRect = firstErrorEl.getBoundingClientRect()
        const targetScrollTop = dialogBody.scrollTop + (errorRect.top - dialogRect.top) - offset
        dialogBody.scrollTo({ top: Math.max(0, targetScrollTop), behavior: 'smooth' })
        return
      }

      // Scroll window to top-most error field - 15px offset
      const errorTop = window.scrollY + firstErrorEl.getBoundingClientRect().top - offset
      window.scrollTo({ top: Math.max(0, errorTop), behavior: 'smooth' })
      return
    }

    // 2. Fallback if no specific error class is found yet: scroll dialog body or window to top + 15px
    const dialogBody = containerElement || (document.querySelector('.el-dialog__body') as HTMLElement | null)
    if (dialogBody && dialogBody.scrollHeight > dialogBody.clientHeight) {
      dialogBody.scrollTo({ top: offset, behavior: 'smooth' })
      return
    }

    window.scrollTo({ top: offset, behavior: 'smooth' })
  }, 350)
}
