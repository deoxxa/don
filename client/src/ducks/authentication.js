// @flow

import axios from 'axios';
import URLSearchParams from 'url-search-params';

export type User = {
  id: string,
  username: string,
  displayName: string,
  avatar: string,
};

export type State = {
  loading: boolean,
  error: ?Error,
  user: ?User,
};

export const authenticationRegister = (
  email: string,
  username: string,
  password: string
) =>
  (dispatch: (a: Object) => void) => {
    const params = new URLSearchParams();
    params.append('email', email);
    params.append('username', username);
    params.append('password', password);

    dispatch(authenticationLoading());

    return axios
      .post('/register', params)
      .then(
        ({ data: { user } }: { data: { user: User } }) =>
          dispatch(authenticationSuccess(user)),
        err => dispatch(authenticationError(err))
      );
  };

export const authenticationLogin = (username: string, password: string) =>
  (dispatch: (a: Object) => void) => {
    const params = new URLSearchParams();
    params.append('username', username);
    params.append('password', password);

    dispatch(authenticationLoading());

    return axios
      .post('/login', params)
      .then(
        ({ data: { user } }: { data: { user: User } }) =>
          dispatch(authenticationSuccess(user)),
        err => dispatch(authenticationError(err))
      );
  };

export const authenticationLogout = () =>
  (dispatch: (a: Object) => void) => {
    dispatch(authenticationLoading());

    return axios
      .post('/logout')
      .then(
        () => dispatch(authenticationReset()),
        err => dispatch(authenticationError(err))
      );
  };

export const authenticationError = (error: Error) => ({
  type: 'don/authentication/ERROR',
  payload: { error },
});
export const authenticationLoading = () => ({
  type: 'don/authentication/LOADING',
  payload: {},
});
export const authenticationSuccess = (user: User) => ({
  type: 'don/authentication/SUCCESS',
  payload: { user },
});
export const authenticationReset = () => ({
  type: 'don/authentication/RESET',
  payload: {},
});

const defaultState = {
  loading: false,
  error: null,
  user: null,
};

export default (
  state: State = defaultState,
  action:
    | { type: 'don/authentication/LOADING', payload: {} }
    | {
        type: 'don/authentication/SUCCESS',
        payload: { user: User },
      }
    | { type: 'don/authentication/ERROR', payload: { error: Error } }
    | { type: 'don/authentication/RESET', payload: {} }
) => {
  switch (action.type) {
    case 'don/authentication/LOADING':
      return {
        ...state,
        loading: true,
        error: null,
      };
    case 'don/authentication/SUCCESS':
      return {
        ...state,
        loading: false,
        user: action.payload.user,
        error: null,
      };
    case 'don/authentication/ERROR':
      return {
        ...state,
        loading: false,
        error: action.payload.error,
      };
    case 'don/authentication/RESET':
      return defaultState;
    default:
      return state;
  }
};
