const ERROR = 'don/publicTimeline/ERROR';
const LOADED = 'don/publicTimeline/LOADED';
const LOADING = 'don/publicTimeline/LOADING';

export const publicTimelineError = error => ({
  type: ERROR,
  payload: { error },
});
export const publicTimelineLoaded = posts => ({
  type: LOADED,
  payload: { posts },
});
export const publicTimelineLoading = () => ({ type: LOADING });

const defaultState = {
  loading: false,
  posts: null,
  error: null,
};

export default (state = defaultState, action) => {
  const { type, payload } = action;

  switch (type) {
    case ERROR:
      return { ...state, loading: false, error: payload.error };
    case LOADED:
      return { ...state, loading: false, error: null, posts: payload.posts };
    case LOADING:
      return { ...state, loading: true, error: null };
    default:
      return state;
  }
};
