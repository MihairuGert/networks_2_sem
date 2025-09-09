package sender;

import java.io.IOException;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketAddress;
import java.io.*;
import java.util.Scanner;

public class Client {

    private final String path;
    private Socket socket;

    public Client(String path) {
        this.path = path;
    }

    private void connect(String ip, int port) {
        try {
            socket = new Socket(InetAddress.getByName(ip), port);
            BufferedReader in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
            BufferedWriter out = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream()));

            System.out.println("Вы что-то хотели сказать? Введите это здесь:");
            Scanner scanner = new Scanner(System.in);
            String word = scanner.next();

            out.write(word + "\n");
            out.flush();
            String serverWord = in.readLine();
            System.out.println(serverWord);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    public void startSend() {

    }

    public static void main(String[] args) {
        Client client = new Client("hui");
        client.connect("192.168.0.120", 8000);
    }
}
