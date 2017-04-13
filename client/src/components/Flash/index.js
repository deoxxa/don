// @flow

import classNames from 'classnames/bind';
import React from 'react';

import Inner from 'components/Inner';

import styles from './styles.css';

const cx = classNames.bind(styles);

const Flash = (
  {
    info,
    success,
    warning,
    error,
    children,
    onClick,
  }: {
    info?: boolean,
    success?: boolean,
    warning?: boolean,
    error?: boolean,
    children?: React.Children,
    onClick?: ?(ev: MouseEvent) => void,
  }
) => (
  <div
    className={cx('flash', { info, success, warning, error, onClick })}
    onClick={onClick}
  >
    <Inner>{children}</Inner>
  </div>
);

export default Flash;
