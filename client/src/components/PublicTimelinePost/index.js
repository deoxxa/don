import React, { Component } from 'react';

import styles from './styles.css';

export default class PublicTimelinePost extends Component {
  render() {
    const { post } = this.props;

    return (
      <div>
        <h3>
          {post.authorName
            ? <span>
                {post.authorName} ({post.authorAcct || 'mystery account'})
              </span>
            : <span>No Username Available</span>}
        </h3>

        <h4><time>{post.time}</time></h4>

        <div
          className={styles.content}
          dangerouslySetInnerHTML={{ __html: post.contentHTML }}
        />
      </div>
    );
  }
}
