package lab3;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.MalformedURLException;
import java.net.URL;
import java.text.DecimalFormat;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

public class HTTPRequester {

    private static final String GRAPH_HOPPER_API_KEY = "";
    private static final String OPEN_WEATHER_API_KEY = "";
    private static final String OPEN_TRIP_MAP_API_KEY = "";

    public String getWeather(Point point) throws IOException {
        return getJSON(String.format("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&units=metric&appid=%s", point.lat, point.lon, OPEN_WEATHER_API_KEY));
    }

    public String getInterestingPlaces(Point point) throws IOException {
        return getJSON(String.format(Locale.US, "http://api.opentripmap.com/0.1/ru/places/radius?radius=3000&lon=%f&lat=%f&format=json&apikey=%s", point.lon, point.lat, OPEN_TRIP_MAP_API_KEY));
    }

    public String getDescription(String xid) throws IOException {
        return getJSON(String.format("http://api.opentripmap.com/0.1/ru/places/xid/%s?apikey=%s", xid, OPEN_TRIP_MAP_API_KEY));
    }

    public CompletableFuture<String> getRequestChain() {
        return CompletableFuture.supplyAsync(this::getLocationPoint)
                .thenCompose(point -> getWeatherAsync(point)
                        .thenCombine(getPlacesAsync(point), this::combineResults));
    }

    private Point getLocationPoint() {
        System.out.println("Search for: ");
        String locationName = getLocationName();
        String response = getJSON(String.format("https://graphhopper.com/api/1/geocode?q=%s&locale=en&key=%s", locationName, GRAPH_HOPPER_API_KEY));
        JSONObject jsonObject = new JSONObject(response);
        JSONArray jsonArray = jsonObject.getJSONArray("hits");
        for (int i = 0; i < jsonArray.length(); i++) {
            JSONObject object = jsonArray.getJSONObject(i);
            System.out.printf("%d) %s[%s]\t%s\n", i+1, object.getString("country"), object.getString("countrycode"), object.getString("name"));
        }
        int chosenVariant = getVariant();
        if (chosenVariant < 1 || chosenVariant > jsonArray.length()) {
            throw new RuntimeException("wrong variant: try agian from 1 to " + jsonArray.length());
        }
        JSONObject object = jsonArray.getJSONObject(chosenVariant - 1);
        JSONObject object1 = object.getJSONObject("point");

        double lat =  object1.getDouble("lat");
        double lon = object1.getDouble("lng");
        return new Point(lon, lat);
    }

    private CompletableFuture<String> getWeatherAsync(Point point) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String weatherJSON = getWeather(point);
                JSONObject jsonObject = new JSONObject(weatherJSON);
                JSONArray weather = jsonObject.getJSONArray("weather");
                JSONObject main = jsonObject.getJSONObject("main");
                DecimalFormat decimalFormat = new DecimalFormat("0.00");
                return String.format("Weather: %s | Temp: %s℃ | Feels like: %s℃ | Pressure: %d hPa",
                        weather.getJSONObject(0).getString("main"),
                        decimalFormat.format(main.getDouble("temp")),
                        decimalFormat.format(main.getDouble("feels_like")),
                        main.getInt("pressure"));
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        });
    }

    private CompletableFuture<String> getPlacesAsync(Point point) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String interestingPlacesJSON = getInterestingPlaces(point);
                ArrayList<LocationData> locationDatum = new ArrayList<>();
                JSONArray places = new JSONArray(interestingPlacesJSON);

                int maxPlaces = Math.min(places.length(), 8);
                List<CompletableFuture<String>> placeFutures = new ArrayList<>();

                for (int i = 0; i < maxPlaces; i++) {
                    JSONObject place = places.getJSONObject(i);
                    String placeId = place.getString("xid");

                    CompletableFuture<String> placeFuture = CompletableFuture.supplyAsync(() -> {
                        try {
                            String placeDetailsJSON = getDescription(placeId);
                            return placeDetailsJSON;
                        } catch (IOException e) {
                            throw new RuntimeException("Failed to get details for place: " + placeId, e);
                        }
                    });

                    placeFutures.add(placeFuture);
                }

                return CompletableFuture.allOf(placeFutures.toArray(new CompletableFuture[0]))
                        .thenApply(v -> {
                            List<String> locations = placeFutures.stream()
                                    .map(CompletableFuture::join)
                                    .toList();
                            StringBuilder stringBuilder = new StringBuilder();
                            for (String s : locations) {
                                JSONObject desc = new JSONObject(s);
                                stringBuilder.append("Name: ").append(desc.getString("name")).append('\n')
                                        .append("Kinds: ").append(desc.getString("kinds")).append('\n');
                                if (!desc.isNull("wikipedia_extracts")) {
                                    stringBuilder.append("Description: ").append(desc.getJSONObject("wikipedia_extracts").getString("text")).append('\n');
                                }
                                stringBuilder.append("\n");
                            }
                            return stringBuilder.toString();
                        })
                        .join();
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        });
    }

    private String combineResults(String weather, String places) {
        return weather + "\n\n" + places;
    }


    private static String getJSON(String URL) {
        StringBuilder responseBuilder = new StringBuilder();
        try {
            URL url = new URL(URL);
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("GET");

            BufferedReader in = new BufferedReader(new InputStreamReader(conn.getInputStream()));
            String inputLine;

            while ((inputLine = in.readLine()) != null) {
                responseBuilder.append(inputLine);
            }
            in.close();
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
        return responseBuilder.toString();
    }

    private String getLocationName() {
        Scanner scanner = new Scanner(System.in);
        return scanner.nextLine();
    }

    private int getVariant() {
        Scanner scanner = new Scanner(System.in);
        return scanner.nextInt();
    }
}
