// @flow

import axios from 'axios';

export type Post = {
  id: string,
  authorName: ?string,
  authorAcct: ?string,
  time: string,
  contentHTML: string,
};

export type State = {
  loading: boolean,
  posts: ?Array<Post>,
  error: ?Error,
};

export const publicTimelineError = (error: Error) => ({
  type: 'don/publicTimeline/ERROR',
  payload: { error },
});
export const publicTimelineLoaded = (posts: Array<Post>) => ({
  type: 'don/publicTimeline/LOADED',
  payload: { posts },
});
export const publicTimelineLoading = () => ({
  type: 'don/publicTimeline/LOADING',
  payload: {},
});

export const publicTimelineFetch = () =>
  (dispatch: (a: Object) => void) => {
    dispatch(publicTimelineLoading());

    return axios
      .get('/')
      .then(
        ({ data: { publicTimeline: { posts } } }) =>
          dispatch(publicTimelineLoaded(posts)),
        error => dispatch(publicTimelineError(error))
      );
  };

const defaultState = {
  loading: false,
  posts: null,
  error: null,
};

export default (
  state: State = defaultState,
  action:
    | { type: 'don/publicTimeline/ERROR', payload: { error: Error } }
    | { type: 'don/publicTimeline/LOADED', payload: { posts: Array<Post> } }
    | { type: 'don/publicTimeline/LOADING', payload: {} }
) => {
  switch (action.type) {
    case 'don/publicTimeline/ERROR':
      return {
        ...state,
        loading: false,
        error: action.payload.error,
      };
    case 'don/publicTimeline/LOADED':
      return {
        ...state,
        loading: false,
        error: null,
        posts: action.payload.posts,
      };
    case 'don/publicTimeline/LOADING':
      return {
        ...state,
        loading: true,
        error: null,
      };
    default:
      return state;
  }
};
