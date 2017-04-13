// @flow

import classNames from 'classnames';
import React from 'react';

import styles from './styles.css';

const Inner = (
  { className, children }: { className?: ?string, children?: React.Children }
) => <div className={classNames(styles.inner, className)}>{children}</div>;

export default Inner;
