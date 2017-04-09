// @flow

import React from 'react';
import { Provider } from 'react-redux';
import { IndexRoute, Route } from 'react-router';

import App from 'containers/App';
import Home from 'containers/Home';

function scrollToTop() {
  window.scrollTo(0, 0);
}

const Root = ({ store }) => (
  <Provider store={store}>
    <App>
      <Route exact path="/" component={Home} />
    </App>
  </Provider>
);

export default Root;
