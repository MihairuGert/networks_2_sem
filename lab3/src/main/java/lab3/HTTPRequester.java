package lab3;

import netscape.javascript.JSObject;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.Scanner;

public class HTTPRequester {
    public void doRequestChain() {
        System.out.println("Search for: ");
        String locationName = getLocationName();
        try {
            URL url = new URL("https://graphhopper.com/api/1/geocode?q="+ locationName +"&locale=en&key=");
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("GET");

            int responseCode = conn.getResponseCode();
            System.out.println("Response Code: " + responseCode);

            BufferedReader in = new BufferedReader(new InputStreamReader(conn.getInputStream()));
            String inputLine;
            StringBuilder responseBuilder = new StringBuilder();

            while ((inputLine = in.readLine()) != null) {
                responseBuilder.append(inputLine);
            }
            in.close();

            String response = responseBuilder.toString();
            System.out.println(response);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }

    }

    private String[] parseJson(String json) {
        String[] locations = null;

        return locations;
    }

    private String getLocationName() {
        Scanner scanner = new Scanner(System.in);
        return scanner.nextLine();
    }
}
