package lab3;

import java.util.concurrent.CompletableFuture;

public class Main {
    public static void main(String[] args) {
        HTTPRequester httpRequester = new HTTPRequester();
        try {
            CompletableFuture<String> future = httpRequester.getRequestChain();
            System.out.println(future.get());
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}