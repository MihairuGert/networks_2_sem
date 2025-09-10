package sender;

public class ClientStatistics {
    private final long timeConnected;
    private final long period;
    private long bytesReceived = 0;
    private long bytesReceivedPeriodAgo;
    private String filename;

    ClientStatistics(long period) {
        timeConnected = System.currentTimeMillis();
        this.period = period;
    }

    public void setBytesReceivedPeriodAgo() {
        bytesReceivedPeriodAgo = bytesReceived;
    }

    public double getInstantSpeed() {
        return (double) (bytesReceived - bytesReceivedPeriodAgo) / period;
    }

    public double getAverageSpeed() {
        return (double) bytesReceived / (System.currentTimeMillis() - timeConnected);
    }

    public void setFilename(String filename) {
        this.filename = filename;
    }

    public String getFilename() {
        return filename;
    }

    public void setBytesReceived(long bytesReceived) {
        this.bytesReceived = bytesReceived;
    }

    public void addBytesReceived(int symRead) {
        bytesReceived += symRead;
    }
}
