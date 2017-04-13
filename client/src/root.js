// @flow

import React from 'react';
import { Provider } from 'react-redux';
import { Route, Switch } from 'react-router';

import App from 'containers/App';
import Home from 'containers/Home';
import Login from 'containers/Login';
import Logout from 'containers/Logout';
import Register from 'containers/Register';

const Root = ({ store }: { store: Object }) => (
  <Provider store={store}>
    <App>
      <Switch>
        <Route exact path="/" component={Home} />
        <Route path="/login" component={Login} />
        <Route path="/logout" component={Logout} />
        <Route path="/register" component={Register} />
      </Switch>
    </App>
  </Provider>
);

export default Root;
