const initialValue = {};

export function defaultCompare(a, b) {
  if (a < b) {
    return -1;
  }

  if (a > b) {
    return 1;
  }

  return 0;
}

export function reverseCompare(a, b) {
  if (a < b) {
    return 1;
  }

  if (a > b) {
    return -1;
  }

  return 0;
}

export default function sort(arr, key, compare = defaultCompare) {
  let last = initialValue;

  let sorted = true;
  for (let i = 0; i < arr.length; i++) {
    if (last !== initialValue && last < arr[i][key]) {
      sorted = false;
      break;
    }

    last = arr[i][key];
  }

  if (sorted) {
    return arr;
  }

  return arr.slice().sort((a, b) => compare(a[key], b[key]));
}
