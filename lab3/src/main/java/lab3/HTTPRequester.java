package lab3;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.Locale;
import java.util.Scanner;

public class HTTPRequester {

    private static final String GRAPH_HOPPER_API_KEY = "";
    private static final String OPEN_WEATHER_API_KEY = "";
    private static final String OPEN_TRIP_MAP_API_KEY = "";

    public String getWeather(double lat, double lon) throws IOException {
        String response = getJSON(String.format("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s", lat, lon, OPEN_WEATHER_API_KEY));
        System.out.println(response);
        return response;
    }

    public String getInterestingPlaces(double lat, double lon) throws IOException {
        String response = getJSON(String.format(Locale.US, "http://api.opentripmap.com/0.1/en/places/radius?radius=3000&lon=%f&lat=%f&format=json&apikey=%s", lon, lat, OPEN_TRIP_MAP_API_KEY));
        System.out.println(response);
        return response;
    }

    public String getDescription(String xid) throws IOException {
        String response = getJSON(String.format("http://api.opentripmap.com/0.1/en/places/xid/%s?apikey=%s", xid, OPEN_TRIP_MAP_API_KEY));
        System.out.println(response);
        return response;
    }

    public void doRequestChain() throws Exception {
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
            throw new Exception("wrong variant: try agian from 1 to " + jsonArray.length());
        }
        JSONObject object = jsonArray.getJSONObject(chosenVariant);
        JSONObject object1 = object.getJSONObject("point");

        double lat =  object1.getDouble("lat");
        double lon = object1.getDouble("lng");
        String weatherJSON = getWeather(lat, lon);
        System.out.println(weatherJSON);
        String interestingJSON = getInterestingPlaces(lat, lon);
        System.out.println(interestingJSON);
    }

    private static String getJSON(String URL) throws IOException {
        URL url = new URL(URL);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");

        BufferedReader in = new BufferedReader(new InputStreamReader(conn.getInputStream()));
        String inputLine;
        StringBuilder responseBuilder = new StringBuilder();

        while ((inputLine = in.readLine()) != null) {
            responseBuilder.append(inputLine);
        }
        in.close();

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
