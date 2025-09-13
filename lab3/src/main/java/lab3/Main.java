package lab3;

public class Main {
    public static void main(String[] args) {
        HTTPRequester httpRequester = new HTTPRequester();
        try {
            httpRequester.doRequestChain();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}