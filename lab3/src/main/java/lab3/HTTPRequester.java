package lab3;

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
            StringBuilder response = new StringBuilder();

            while ((inputLine = in.readLine()) != null) {
                response.append(inputLine);
            }
            in.close();
            System.out.println(response.toString());
        } catch (IOException e) {
            throw new RuntimeException(e);
        }

    }

    private String getLocationName() {
        Scanner scanner = new Scanner(System.in);
        return scanner.nextLine();
    }
}
