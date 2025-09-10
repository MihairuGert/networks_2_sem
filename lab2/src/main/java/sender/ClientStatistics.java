package sender;

public class ClientStatistics {
    private final long timeConnected;
    private final long period;
    private long bytesReceived = 0;
    private long bytesReceivedPeriodAgo;
    private long fileSize;
    private String filename;

    private double bytesToKb(long bytes) {
        return (double) bytes / 1024;
    }

    private double bytesToMb(long bytes) {
        return (double) bytes / (1024*1024);
    }

    ClientStatistics(long period) {
        timeConnected = System.currentTimeMillis();
        this.period = period;
    }

    public void setBytesReceivedPeriodAgo() {
        bytesReceivedPeriodAgo = bytesReceived;
    }

    public double getInstantSpeed() {
        return bytesToMb(bytesReceived - bytesReceivedPeriodAgo) / period;
    }

    public double getAverageSpeed() {
        return bytesToMb(bytesReceived) / ((double) (System.currentTimeMillis() - timeConnected) /1000);
    }

    public double getPercent() {
        return ((double) bytesReceived * 100) / fileSize;
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

    public void setFileSize(long fileSize) {
        this.fileSize = fileSize;
    }
}
