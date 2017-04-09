import classNames from 'classnames/bind';
import styles from 'font-awesome/css/font-awesome.css';
import React, { Component, PropTypes } from 'react';

const cx = classNames.bind(styles);

function without(o, ...keys) {
  const r = {};

  for (const k in o) {
    if (keys.includes(k)) {
      continue;
    }

    r[k] = o[k];
  }

  return r;
}

export default class FontAwesome extends Component {
  static propTypes = {
    icon: PropTypes.string.isRequired,
    className: PropTypes.string,
  };

  render() {
    return (
      <i
        {...without(this.props, 'className', 'icon')}
        className={cx('fa', `fa-${this.props.icon}`, this.props.className)}
      />
    );
  }
}
