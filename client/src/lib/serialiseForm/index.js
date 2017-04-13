// @flow

// adapted from code in [freiform](https://github.com/mechanoid/freiform) by
// Falk Hoppe (mechanoid), licensed under the Apache License 2.0

function searchParamsFromFormData(data: FormData): string {
  const searchParams = new URLSearchParams();
  data.forEach((val, key) => searchParams.append(key, val));
  return searchParams.toString();
}

export function serialiseForm(form: HTMLFormElement): string {
  const data = new FormData(form);
  const action = new URL(form.action);
  action.search = searchParamsFromFormData(data);
  return action.pathname + action.search;
}
