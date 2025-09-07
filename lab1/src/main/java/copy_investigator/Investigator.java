package copy_investigator;

import java.io.IOException;
import java.net.*;
import java.util.Enumeration;
import java.util.Iterator;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class Investigator {
    private final int port = 1500;
    private MulticastSocket multicastSocket;
    private final String uniqueMsg = "ASK";
    private InetAddress group;

    private final int askInterval = 250;
    private final int askReceiveTimeout = 500;

    private void setMulticastSocket(String string) throws IOException {
        group = InetAddress.getByName(string);
        multicastSocket = new MulticastSocket(port);
        multicastSocket.setSoTimeout(askReceiveTimeout);
        if (group instanceof Inet6Address) {
            multicastSocket.setOption(StandardSocketOptions.IP_MULTICAST_IF, getIPv6Interface());
        }
        multicastSocket.joinGroup(group);
    }

    private NetworkInterface getIPv6Interface() throws IOException {
        Enumeration<NetworkInterface> interfaces = NetworkInterface.getNetworkInterfaces();
        while (interfaces.hasMoreElements()) {
            NetworkInterface ni = interfaces.nextElement();
            if (ni.isUp() && !ni.isLoopback() && ni.supportsMulticast()) {
                Enumeration<InetAddress> addresses = ni.getInetAddresses();
                while (addresses.hasMoreElements()) {
                    if (addresses.nextElement() instanceof Inet6Address) {
                        return ni;
                    }
                }
            }
        }
        throw new IOException("No suitable IPv6 network interface found");
    }

    private void sendMsg() {
        byte[] buffer = uniqueMsg.getBytes();
        try {
            DatagramPacket datagramPacket = new DatagramPacket(buffer, buffer.length, group, port);
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

    Investigator(String ip) {
        try {
            setMulticastSocket(ip);
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

    private ConcurrentHashMap<String, Long> clientDatum = new ConcurrentHashMap<>();
    private volatile boolean wasAdded = false;

    public void startChecking() {
        new Thread(()->{
            while (!multicastSocket.isClosed()) {
                wait_millis(askInterval / 10);
                sendMsg();
            }
        }).start();
        new Thread(()->{
            while (!multicastSocket.isClosed()) {
                ClientData clientData = receiveMsg();
                Long ret = clientDatum.put(processMsg(clientData), System.currentTimeMillis());
                if (ret == null) {
                    wasAdded = true;
                }
            }
        }).start();
        new Thread(()->{
            while (!multicastSocket.isClosed()) {
                wait_millis(askInterval);
                boolean isChanged = cleanMap() || wasAdded;
                wasAdded = false;
                if (!isChanged) {
                    continue;
                }
                System.out.println("<------------>\n");
                clientDatum.forEach((key, value) -> System.out.println(key + '\n'));
                System.out.println("<------------>\n");
            }
        }).start();
    }

    private boolean cleanMap() {
        Iterator<Map.Entry<String, Long>> iterator = clientDatum.entrySet().iterator();
        Long currentTime = System.currentTimeMillis();
        boolean isChanged = false;
        while (iterator.hasNext()) {
            Map.Entry<String, Long> entry = iterator.next();
            if (currentTime - entry.getValue() > 500) {
                iterator.remove();
                isChanged = true;
            }
        }
        return isChanged;
    }
}
