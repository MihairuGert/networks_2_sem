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
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.Locale;
import java.util.Scanner;
import java.util.concurrent.CompletableFuture;

public class HTTPRequester {

    private static final String GRAPH_HOPPER_API_KEY = "";
    private static final String OPEN_WEATHER_API_KEY = "";
    private static final String OPEN_TRIP_MAP_API_KEY = "";

    public String getWeather(Point point) throws IOException {
        return getJSON(String.format("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&units=metric&appid=%s", point.lat, point.lon, OPEN_WEATHER_API_KEY));
    }

    public String getInterestingPlaces(Point point) throws IOException {
        return getJSON(String.format(Locale.US, "http://api.opentripmap.com/0.1/en/places/radius?radius=3000&lon=%f&lat=%f&format=json&apikey=%s", point.lon, point.lat, OPEN_TRIP_MAP_API_KEY));
    }

    public String getDescription(String xid) throws IOException {
        return getJSON(String.format("http://api.opentripmap.com/0.1/en/places/xid/%s?apikey=%s", xid, OPEN_TRIP_MAP_API_KEY));
    }

    public CompletableFuture<String> getRequestChain() throws Exception {
        CompletableFuture<String> getInfo = CompletableFuture.supplyAsync(() -> {
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
        }).thenCompose(point -> {
            CompletableFuture<String> weatherFuture = CompletableFuture.supplyAsync(() -> {
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
//            CompletableFuture<String> placesDescFuture = CompletableFuture.supplyAsync((locationData) -> {
//                //String descJSON = getDescription();
//            });
            CompletableFuture<String> placesFuture = CompletableFuture.supplyAsync(() -> {
                try {
                    String interestingPlacesJSON = getInterestingPlaces(point);
                    ArrayList<LocationData> locationDatum = new ArrayList<>();
                    JSONArray weather = new JSONArray(interestingPlacesJSON);
                    // one more async get query
                    return interestingPlacesJSON;
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            });
            return weatherFuture.thenCombine(placesFuture, (weather, places) -> {
                return weather + "\n" + places;
            });
        });
        return getInfo;
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
