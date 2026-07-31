import type { LabelHTMLAttributes } from 'react'

import styles from './Label.module.css'

export interface LabelProps extends LabelHTMLAttributes<HTMLLabelElement> {
  /**
   * Id of the form control this label describes. Required (rather than
   * optional, as in the DOM) so a <Label> can never be rendered without an
   * associated control — an unassociated label is invisible to assistive
   * technology.
   */
  htmlFor: string
  required?: boolean
}

/**
 * Standalone form label primitive, for composite fields that render their own
 * control (the <Input> primitive already renders its own associated label and
 * does not use this component).
 *
 * When `required` is set, the requirement is conveyed twice: a decorative
 * asterisk for sighted users (aria-hidden, so screen readers do not announce
 * "asterisk") and a visually hidden "(required)" text so assistive technology
 * users get the same information (WCAG 2.1 AA).
 */
export function Label({
  htmlFor,
  required,
  className,
  children,
  ...rest
}: LabelProps) {
  const labelClassNames = [styles.label, className].filter(Boolean).join(' ')

  return (
    <label htmlFor={htmlFor} className={labelClassNames} {...rest}>
      {children}
      {required && (
        <>
          <span aria-hidden="true" className={styles.requiredMark}>
            *
          </span>
          <span className={styles.visuallyHidden}>(required)</span>
        </>
      )}
    </label>
  )
}
