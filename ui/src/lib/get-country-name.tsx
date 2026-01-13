export const getCountryName = (countryCode: string, locale = "en") => {
  try {
    const regionNames = new Intl.DisplayNames([locale], { type: "region" });
    return regionNames.of(countryCode.toUpperCase());
  } catch (error) {
    console.error("Error getting country name:", error);
    return null; // Or a default value
  }
};
