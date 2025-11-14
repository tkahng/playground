export const decodeRedirectTo = (
  redirectTo: string | null,
  defaultTo: string = "/"
): URL => {
  if (!redirectTo?.length) {
    return new URL(defaultTo, window.location.origin);
  }
  return new URL(decodeURIComponent(redirectTo), window.location.origin);
};
