package copy_investigator;

import java.io.IOException;
import java.net.*;
import java.util.concurrent.ConcurrentHashMap;

public class Investigator {
    private final int port = 1500;
    private MulticastSocket multicastSocket;
    private final String uniqueMsg = "ASK";

    private final int askInterval = 250;
    private final int askReceiveTimeout = 500;

    private String getGroupIP() {
        return "239.255.255.250";
    }

    private void sendMsg() {
        byte[] buffer = uniqueMsg.getBytes();
        try {
            DatagramPacket datagramPacket = new DatagramPacket(buffer, buffer.length, InetAddress.getByName(getGroupIP()), port);
            multicastSocket.send(datagramPacket);
        } catch (IOException e) {
            System.out.println(e.getMessage());
        }
    }

    private ClientData receiveMsg() {
        byte[] buffer = new byte[512];
        DatagramPacket responsePacket = new DatagramPacket(buffer, buffer.length);
        try {
            multicastSocket.receive(responsePacket);
        } catch (Exception e) {
            System.out.println(e.getMessage());
        }
        return new ClientData(responsePacket.getAddress(), new String(responsePacket.getData()).trim());
    }

    Investigator() {
        try {
            multicastSocket = new MulticastSocket(port);
            InetAddress group = InetAddress.getByName(getGroupIP());
            multicastSocket.joinGroup(group);
            multicastSocket.setSoTimeout(askReceiveTimeout);
        } catch (Exception e) {
            System.out.println(e.getMessage());
        }
    }

    private void wait_millis(int millis) {
        try {
            Thread.sleep(askInterval);
        } catch (InterruptedException e) {
            System.out.println(e.getMessage());
        }
    }

    private String processMsg(ClientData clientData) {
        if (clientData.getMsg().equals(uniqueMsg)) {
            return clientData.getAddress().toString();
        }
        return "Corrupted message";
    }

    private ConcurrentHashMap<String, Integer> clientDatum = new ConcurrentHashMap<>();

    public void startChecking() {
        new Thread(()->{
            while (!multicastSocket.isClosed()) {
                wait_millis(askInterval/10);
                sendMsg();
            }
        }).start();
        new Thread(()->{
            while (!multicastSocket.isClosed()) {
                ClientData clientData = receiveMsg();
                clientDatum.put(processMsg(clientData), 1);
            }
        }).start();
        new Thread(()->{
            ConcurrentHashMap<String, Integer> prevClientDatum = new ConcurrentHashMap<>();
            while (!multicastSocket.isClosed()) {
                wait_millis(askInterval);
                if (!prevClientDatum.equals(clientDatum)) {
                    prevClientDatum.clear();
                    prevClientDatum.putAll(clientDatum);
                    clientDatum.clear();
                    System.out.println("<------------>\n");
                    prevClientDatum.forEach((key, value) -> System.out.println(key + '\n'));
                    System.out.println("<------------>\n");
                }
            }
        }).start();
    }
}
