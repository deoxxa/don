// @flow

import classNames from 'classnames';
import React from 'react';

import styles from './styles.css';

const AuthenticationForm = (
  {
    className,
    children,
    ...props
  }: { className: string, children?: React.children }
) => (
  <form className={classNames(styles.form, className)} {...props}>
    <fieldset className={styles.fields}>
      {children}
    </fieldset>
  </form>
);

export default AuthenticationForm;
